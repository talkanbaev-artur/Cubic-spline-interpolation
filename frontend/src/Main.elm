module Main exposing (main)

import Browser
import Html exposing (..)
import Html.Attributes exposing (type_, value)
import Html.Events exposing (onClick, onInput)
import Http
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
    { fValsData : List ( Float, Float ) -- f(x) evaluated function on 1/10000 rate
    , sValsDataSecondDer : SplineData -- S_3(x,f^h) for f''  in M calc
    , sValsDataFirstDer : SplineData -- S_3(x, f^h) for f' in M Calc
    , sValsDataNormal : SplineData -- S_3(x,f^h) for M_1 = 0 and M_n = 0 (normal spline)
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



-- UPDATE


type Msg
    = None


update : Msg -> Model -> ( Model, Cmd Msg )
update msg model =
    ( model, Cmd.none )



-- VIEW


view : Model -> Html.Html Msg
view model =
    div []
        [ h1 []
            [ text "Cubic spline interpolation calculator" ]
        , div []
            [ p [] [ text <| String.fromInt model.options.gridPointsCoefficent ]
            , p [] [ text <| String.fromInt model.options.epsilionCoef ]
            ]
        ]
