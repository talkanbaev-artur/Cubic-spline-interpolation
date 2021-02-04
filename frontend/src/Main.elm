module Main exposing (main)

import Browser
import Graph exposing (convertPoints, viewGraph)
import Html exposing (..)
import Html.Attributes exposing (type_, value)
import Html.Events exposing (on, onClick, onInput, targetValue)
import Http
import Json.Decode exposing (float, list)
import Json.Decode.Pipeline exposing (required)
import Json.Encode as Enc
import Platform.Cmd exposing (Cmd)


main : Program () Model Msg
main =
    Browser.element
        { init = \_ -> ( initModel, Cmd.none )
        , subscriptions = \_ -> Sub.none
        , view = view
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
    { fValsX : List Float -- x on 1/10000 rate
    , fValsY : List Float

    --, sValsDataSecondDer : SplineData -- S_3(x,f^h) for f''  in M calc
    --, sValsDataFirstDer : SplineData -- S_3(x, f^h) for f' in M Calc
    --, sValsDataNormal : SplineData -- S_3(x,f^h) for M_1 = 0 and M_n = 0 (normal spline)
    }


type alias SplineData =
    { xVals : List Float -- x vals
    , yVals : List Float -- f(x) vals
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


getMinMax list =
    let
        max =
            case List.maximum list of
                Just val ->
                    val + 0.1

                Nothing ->
                    1

        min =
            case List.minimum list of
                Just val ->
                    val

                Nothing ->
                    0
    in
    ( min, max )



-- UPDATE


type Msg
    = IncrementedGridCoef
    | DecrementedGridCoef
    | IncEpsCoef
    | DecEpsCoef
    | ChangedFunc String
    | ClickedDraw
    | GotGraph (Result Http.Error EvaluatedFunction)


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
    Json.Decode.succeed EvaluatedFunction
        |> required "f_x" (list Json.Decode.float)
        |> required "f_h" (list Json.Decode.float)


optionsEncode : RequestOptions -> Enc.Value
optionsEncode options =
    Enc.object
        [ ( "points", Enc.int <| 1 + (2 ^ options.gridPointsCoefficent) )
        , ( "eps", Enc.float (2.0 ^ (toFloat <| negate options.epsilionCoef)) )
        , ( "func", Enc.string <| funcToShortString options.function )
        ]



-- VIEW


view : Model -> Html.Html Msg
view model =
    div []
        [ viewUnloaded model
        , case model.state of
            Ready ->
                viewUnloadedGraph

            Loading ->
                viewUnloadedGraph

            Loaded data ->
                viewGraph <| ( convertPoints data.fValsX data.fValsY, getMinMax data.fValsX, getMinMax data.fValsY )

            Errored _ ->
                viewUnloadedGraph
        ]


viewUnloadedGraph =
    div [] [ text "Graph is not loaded yet" ]


viewUnloaded : Model -> Html.Html Msg
viewUnloaded model =
    div []
        [ h1 []
            [ text "Cubic spline interpolation calculator new" ]
        , div []
            [ p [] [ text <| String.fromInt model.options.gridPointsCoefficent ]
            , p [] [ text <| String.fromInt model.options.epsilionCoef ]
            , p [] [ text <| funcToString model.options.function ]
            ]
        , div []
            [ div [] [ button [ onClick IncrementedGridCoef ] [ text "+" ], text "Grid handler", button [ onClick DecrementedGridCoef ] [ text "-" ] ]
            , div [] [ button [ onClick IncEpsCoef ] [ text "+" ], text "Eps handler", button [ onClick DecEpsCoef ] [ text "-" ] ]
            , div []
                [ select [ onChangeSelect ChangedFunc ]
                    [ viewFuncOption FirstTest
                    , viewFuncOption SecondTest
                    , viewFuncOption DeltaRegTest
                    , viewFuncOption SmoothStepTest
                    , viewFuncOption PersonalTest
                    ]
                ]
            , div []
                [ button [ onClick ClickedDraw ] [ text "Draw" ]
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
