module Main exposing (main)

import Browser
import Color
import Css exposing (..)
import DoubleSlider
import Graph exposing (convertPoints, viewGraph)
import Html
import Html.Styled exposing (..)
import Html.Styled.Attributes exposing (..)
import Html.Styled.Events exposing (on, onClick, onInput, targetValue)
import Http
import Json.Decode exposing (Decoder, float, list)
import Json.Decode.Pipeline as JDP exposing (required)
import Json.Encode as Enc
import Platform.Cmd exposing (Cmd)


main : Program () Model Msg
main =
    Browser.element
        { init = \_ -> ( initModel, Cmd.none )
        , subscriptions = \_ -> Sub.none
        , view = view >> Html.Styled.toUnstyled
        , update = update
        }



-- The problem type decribes potential issues with the http load request


type Problem
    = InvalidData
    | ServerIsDown
    | InternalError String



-- The status is used for pattern matching on the app's state


type Status
    = Ready -- The page is ready for test data selection
    | Loading -- Evalution period
    | Loaded EvaluatedFunction -- the data is loaded
    | Errored Problem -- error is present


type alias Model =
    { state : Status
    , options : RequestOptions
    }


type alias RequestOptions =
    { gridPointsCoefficent : Int -- Coeficient which determines the number of x vals for spline construction
    , epsilionCoef : Int -- determines the size of the epsilion value in grid function calculations
    , function : TestFunction -- determines the function for calculations
    }


type TestFunction
    = FirstTest -- boundry near x=0 function
    | SecondTest -- boundry near x=0 and x=1
    | DeltaRegTest -- delta-function regularisation
    | SmoothStepTest -- smooth-stepped function
    | PersonalTest -- personal example


type alias EvaluatedFunction =
    { fVals : List ( Float, Float ) -- x on 1/10000 rate
    , sValsDataFirstDer : SplineData -- S_3(x, f^h) for f' in M Calc
    , sValsDataSecondDer : SplineData -- S_3(x,f^h) for f''  in M calc
    , bounds : RangeControl

    --, sValsDataNormal : SplineData -- S_3(x,f^h) for M_1 = 0 and M_n = 0 (normal spline)
    }


type alias RangeControl =
    { slider : DoubleSlider.DoubleSlider Msg
    , bounds : ( Float, Float )
    }


type alias SplineData =
    { xVals : List ( Float, Float ) -- x vals
    , err : Float -- Max Err(f) on y control grid
    }


initModel : Model
initModel =
    { state = Ready
    , options = initOptions
    }


initOptions : RequestOptions
initOptions =
    { gridPointsCoefficent = 2
    , epsilionCoef = 0
    , function = FirstTest
    }



-- UPDATE


type Msg
    = IncrementedGridCoef
    | DecrementedGridCoef
    | IncEpsCoef
    | DecEpsCoef
    | ChangedFunc String
    | ClickedDraw
    | GotGraph (Result Http.Error EvaluatedFunction)
    | SliderLowChange Float
    | SliderHighChange Float


update : Msg -> Model -> ( Model, Cmd Msg )
update msg ({ options } as model) =
    case msg of
        IncrementedGridCoef ->
            let
                newVal =
                    if options.gridPointsCoefficent + 1 > 16 then
                        16

                    else
                        options.gridPointsCoefficent + 1
            in
            ( { model | options = { options | gridPointsCoefficent = newVal } }
            , Cmd.none
            )

        DecrementedGridCoef ->
            let
                newVal =
                    if options.gridPointsCoefficent - 1 < 2 then
                        2

                    else
                        options.gridPointsCoefficent - 1
            in
            ( { model | options = { options | gridPointsCoefficent = newVal } }
            , Cmd.none
            )

        IncEpsCoef ->
            let
                newVal =
                    if options.epsilionCoef + 1 > 16 then
                        16

                    else
                        options.epsilionCoef + 1
            in
            ( { model | options = { options | epsilionCoef = newVal } }
            , Cmd.none
            )

        DecEpsCoef ->
            let
                newVal =
                    if options.epsilionCoef - 1 < 0 then
                        0

                    else
                        options.epsilionCoef - 1
            in
            ( { model | options = { options | epsilionCoef = newVal } }
            , Cmd.none
            )

        ChangedFunc f ->
            let
                newFunction =
                    case f of
                        "first" ->
                            FirstTest

                        "second" ->
                            SecondTest

                        "delta" ->
                            DeltaRegTest

                        "smooth" ->
                            SmoothStepTest

                        "custom" ->
                            PersonalTest

                        _ ->
                            FirstTest
            in
            ( { model | options = { options | function = newFunction } }
            , Cmd.none
            )

        ClickedDraw ->
            ( model, evaluateFuncOnServer model.options )

        GotGraph (Ok eval) ->
            ( { model | state = Loaded eval }
            , Cmd.none
            )

        GotGraph (Err _) ->
            ( { model | state = Errored InvalidData }
            , Cmd.none
            )

        SliderLowChange val ->
            case model.state of
                Loaded ({ bounds } as data) ->
                    let
                        newSlider =
                            DoubleSlider.updateLowValue val data.bounds.slider

                        newBounds =
                            { bounds | slider = newSlider, bounds = Tuple.pair val <| Tuple.second data.bounds.bounds }

                        newEval =
                            { data | bounds = newBounds }
                    in
                    ( { model | state = Loaded newEval }, Cmd.none )

                _ ->
                    ( model, Cmd.none )

        SliderHighChange val ->
            case model.state of
                Loaded ({ bounds } as data) ->
                    let
                        newSlider =
                            DoubleSlider.updateHighValue val bounds.slider

                        newBounds =
                            { bounds | slider = newSlider, bounds = Tuple.pair (Tuple.first bounds.bounds) val }

                        newEval =
                            { data | bounds = newBounds }
                    in
                    ( { model | state = Loaded newEval }, Cmd.none )

                _ ->
                    ( model, Cmd.none )



-- HTTP


evaluateFuncOnServer : RequestOptions -> Cmd Msg
evaluateFuncOnServer options =
    Http.post
        { url = "http://localhost:8000/api/spline"
        , body = Http.jsonBody <| optionsEncode options
        , expect = Http.expectJson GotGraph graphDecoder
        }


graphDecoder : Json.Decode.Decoder EvaluatedFunction
graphDecoder =
    let
        createPairVals : List Float -> List Float -> List ( Float, Float )
        createPairVals xs ys =
            List.map2 Tuple.pair xs ys

        slider =
            { slider =
                DoubleSlider.init
                    { min = 0
                    , max = 1
                    , lowValue = 0
                    , highValue = 1
                    , step = 0.05
                    , onLowChange = SliderLowChange
                    , onHighChange = SliderHighChange
                    }
            , bounds = ( 0.0, 1.0 )
            }

        toDecoder : List Float -> List Float -> SplineData -> SplineData -> Decoder EvaluatedFunction
        toDecoder f_x f_y firstDev secondDer =
            Json.Decode.succeed (EvaluatedFunction (createPairVals f_x f_y) firstDev secondDer slider)
    in
    Json.Decode.succeed toDecoder
        |> JDP.required "f_x" (Json.Decode.list Json.Decode.float)
        |> JDP.required "f_h" (Json.Decode.list Json.Decode.float)
        |> JDP.required "first_spline" decodeSplineJSON
        |> JDP.required "second_spline" decodeSplineJSON
        |> JDP.resolve


decodeSplineJSON =
    let
        toDecoderSplineData : List Float -> List Float -> Float -> Decoder SplineData
        toDecoderSplineData xs ys err =
            Json.Decode.succeed (SplineData (List.map2 Tuple.pair xs ys) err)
    in
    Json.Decode.succeed toDecoderSplineData
        |> JDP.required "x" (Json.Decode.list Json.Decode.float)
        |> JDP.required "y" (Json.Decode.list Json.Decode.float)
        |> JDP.required "err" Json.Decode.float
        |> JDP.resolve


optionsEncode : RequestOptions -> Enc.Value
optionsEncode options =
    Enc.object
        [ ( "points", Enc.int <| 1 + (2 ^ options.gridPointsCoefficent) )
        , ( "eps", Enc.float (2.0 ^ (toFloat <| negate options.epsilionCoef)) )
        , ( "func", Enc.string <| funcToShortString options.function )
        ]



-- VIEW


view : Model -> Html Msg
view model =
    div
        [ css
            [ displayFlex
            , minHeight (vh 100)
            , backgroundColor
                (rgb 240 240 240)
            ]
        ]
        [ viewUnloaded model
        , case model.state of
            Ready ->
                viewUnloadedGraph

            Loading ->
                viewUnloadedGraph

            Loaded data ->
                div
                    [ css
                        [ flexGrow (num 5)
                        , margin auto
                        ]
                    ]
                    [ p [ css [ padding (rem 1) ] ] [ text <| "The error is: " ++ String.fromFloat data.sValsDataSecondDer.err ]
                    , Html.Styled.fromUnstyled (viewGraph (renderData data) data.bounds.bounds)
                    , div [ css [ padding2 (rem 0) (rem 1) ] ] [ Html.Styled.fromUnstyled <| DoubleSlider.view data.bounds.slider ]
                    ]

            Errored _ ->
                viewUnloadedGraph
        ]


renderData : EvaluatedFunction -> List ( List ( Float, Float ), Color.Color )
renderData data =
    [ ( data.sValsDataFirstDer.xVals, Color.black )
    , ( data.sValsDataSecondDer.xVals, Color.blue )
    , ( data.fVals, Color.red )
    ]


viewUnloadedGraph =
    div
        [ css
            [ displayFlex
            , margin auto
            ]
        ]
        [ p
            [ css
                [ fontSize (px 36)
                ]
            ]
            [ text "Graph is not loaded yet" ]
        ]


viewUnloaded : Model -> Html Msg
viewUnloaded model =
    div
        [ css
            [ padding (rem 1)
            , borderRight3 (px 1) solid (rgb 200 200 200)
            ]
        ]
        [ h1
            [ css
                [ fontSize (px 20)
                ]
            ]
            [ text "Cubic spline interpolation calculator" ]
        , div
            [ css [ margin auto, displayFlex, flexDirection column, padding2 (px 0) (rem 0.75) ]
            ]
            [ p [] [ text <| "Number of points: 1 + 2", sup [] [ text <| String.fromInt model.options.gridPointsCoefficent ], text <| " = " ++ (String.fromInt <| 1 + 2 ^ model.options.gridPointsCoefficent) ]
            , p [] [ text <| "The epsilon value is: 2", sup [] [ text <| String.fromInt (negate model.options.epsilionCoef) ], text <| " = " ++ (String.fromFloat <| 2.0 ^ (toFloat <| negate model.options.epsilionCoef)) ]
            ]
        , div []
            [ viewPlusMinController DecrementedGridCoef "Grid handler" IncrementedGridCoef
            , viewPlusMinController DecEpsCoef "Epsilon handler" IncEpsCoef
            , div []
                [ select
                    [ onChangeSelect ChangedFunc
                    , css
                        [ margin (px 10)
                        ]
                    ]
                    [ viewFuncOption FirstTest
                    , viewFuncOption SecondTest
                    , viewFuncOption DeltaRegTest
                    , viewFuncOption SmoothStepTest
                    , viewFuncOption PersonalTest
                    ]
                ]
            , div [ css [ padding (rem 0.75) ] ]
                [ button
                    [ onClick ClickedDraw
                    , css [ padding (rem 0.5), backgroundColor (rgb 200 200 200), borderRadius (rem 0.25), borderStyle solid, border (px 2) ]
                    ]
                    [ text "Draw" ]
                ]
            ]
        ]


funcToString : TestFunction -> String
funcToString f =
    case f of
        FirstTest ->
            "First simple function"

        SecondTest ->
            "Second simple function"

        DeltaRegTest ->
            "Delta regression function"

        SmoothStepTest ->
            "Smooth step function"

        PersonalTest ->
            "Custom function"


funcToShortString : TestFunction -> String
funcToShortString f =
    case f of
        FirstTest ->
            "first"

        SecondTest ->
            "second"

        DeltaRegTest ->
            "delta"

        SmoothStepTest ->
            "smooth"

        PersonalTest ->
            "custom"


viewFuncOption : TestFunction -> Html Msg
viewFuncOption f =
    option
        [ value <| funcToShortString f ]
        [ text <| funcToString f ]


onChangeSelect : (String -> msg) -> Attribute msg
onChangeSelect msg =
    on "change" <| Json.Decode.map msg targetValue


viewPlusMinController decmsg cntName incmsg =
    div
        [ css
            [ padding2 (rem 0.5) (rem 0.75)
            , margin auto
            , displayFlex
            , justifyContent spaceBetween
            ]
        ]
        [ button [ onClick decmsg ] [ text "-" ]
        , text cntName
        , button [ onClick incmsg ] [ text "+" ]
        ]
