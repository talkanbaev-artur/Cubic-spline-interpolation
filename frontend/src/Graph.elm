module Graph exposing (basic, convertPoints, main, viewGraph)

import Color exposing (Color)
import Html
import Path exposing (Path)
import Scale exposing (ContinuousScale)
import Scale.Color
import Shape
import SubPath exposing (SubPath)
import TypedSvg exposing (g, line, rect, svg, text_)
import TypedSvg.Attributes as Explicit exposing (fill, fontFamily, stroke, transform, viewBox)
import TypedSvg.Attributes.InPx exposing (height, strokeWidth, width, x, x1, x2, y, y1, y2)
import TypedSvg.Core exposing (Svg, text)
import TypedSvg.Types exposing (Paint(..), Transform(..), percent)


w : Float
w =
    990


h : Float
h =
    450


padding : Float
padding =
    50


pointsD : List ( Float, Float )
pointsD =
    [ ( 0.1, 1.1 )
    , ( 0.2, 0.6 )
    , ( 0.35, 0.3 )
    , ( 0.45, 0.3 )
    , ( 0.6, 0.2 )
    , ( 0.9, 0.8 )
    , ( 1.2, 0.6 )
    , ( 1.5, 0.9 )
    , ( 1.7, 0.2 )
    , ( 1.9, 0.1 )
    ]


xScale : ContinuousScale Float
xScale =
    Scale.linear ( padding, w - padding ) ( 0, 2 )


yScale : ContinuousScale Float
yScale =
    Scale.linear ( h - padding, padding ) ( 0, 1 )


preparedPoints : List ( Float, Float ) -> List ( Float, Float )
preparedPoints points =
    List.map (\( x, y ) -> ( Scale.convert xScale x, Scale.convert yScale y )) points


xGridLine : Int -> Float -> Svg msg
xGridLine index tick =
    line
        [ y1 0
        , Explicit.y2 (percent 100)
        , x1 (Scale.convert xScale tick)
        , x2 (Scale.convert xScale tick)
        , stroke <| Paint Color.white
        , strokeWidth (Basics.max (toFloat (modBy 2 index)) 0.5)
        ]
        []


yGridLine : Int -> Float -> Svg msg
yGridLine index tick =
    line
        [ x1 0
        , Explicit.x2 (percent 100)
        , y1 (Scale.convert yScale tick)
        , y2 (Scale.convert yScale tick)
        , stroke <| Paint Color.white
        , strokeWidth (Basics.max (toFloat (modBy 2 index)) 0.5)
        ]
        []


type alias Curve =
    List ( Float, Float ) -> SubPath


drawCurve : ( List ( Float, Float ), Curve, Color ) -> Svg msg
drawCurve ( points, curve, color ) =
    List.map Just (preparedPoints points)
        |> Shape.line curve
        |> (\path -> Path.element path [ stroke (Paint color), fill PaintNone, strokeWidth 2 ])


viewGraph model =
    Html.div []
        [ svg [ viewBox 0 0 w h ]
            [ rect [ width w, height h, fill <| Paint <| Color.rgb255 223 223 223 ] []
            , g [] <| List.indexedMap yGridLine <| Scale.ticks yScale 10
            , g [] <| List.indexedMap xGridLine <| Scale.ticks xScale 10
            , g [] <|
                List.map drawCurve model
            ]
        ]


basic : List ( Float, Float ) -> List ( List ( Float, Float ), Curve, Color )
basic prefix =
    [ ( prefix, Shape.linearCurve, Color.black ) ]


convertPoints : List Float -> List Float -> List ( Float, Float )
convertPoints xs ys =
    List.map2 Tuple.pair xs ys


main =
    viewGraph <| basic pointsD
