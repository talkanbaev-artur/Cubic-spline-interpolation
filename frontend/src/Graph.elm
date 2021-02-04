module Graph exposing (convertPoints, viewGraph)

import Axis
import Color exposing (Color)
import Html
import Html.Attributes exposing (class)
import Path exposing (Path)
import Scale exposing (ContinuousScale)
import Scale.Color
import Shape
import SubPath exposing (SubPath)
import TypedSvg exposing (g, line, svg)
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


xScale : ( Float, Float ) -> ContinuousScale Float
xScale minMaxV =
    Scale.linear ( padding, w - padding ) minMaxV


yScale : ( Float, Float ) -> ContinuousScale Float
yScale minMaxV =
    Scale.linear ( h - padding, padding ) minMaxV


preparedPoints : ContinuousScale Float -> ContinuousScale Float -> List ( Float, Float ) -> List ( Float, Float )
preparedPoints xSc ySc points =
    List.map (\( x, y ) -> ( Scale.convert xSc x, Scale.convert ySc y )) points


type alias Curve =
    List ( Float, Float ) -> SubPath


drawCurve : ( List ( Float, Float ), ( Float, Float ), ( Float, Float ) ) -> Html.Html msg
drawCurve ( points, xminmax, yminmax ) =
    List.map Just (preparedPoints (xScale xminmax) (yScale yminmax) points)
        |> Shape.line Shape.linearCurve
        |> (\path -> Path.element path [ stroke (Paint Color.black), fill PaintNone, strokeWidth 2 ])


viewGraph : ( List ( Float, Float ), ( Float, Float ), ( Float, Float ) ) -> Html.Html msg
viewGraph ( points, x, y ) =
    Html.div []
        [ svg [ viewBox 0 0 w h ]
            [ g [ transform [ Translate padding 0 ] ] [ Axis.left [ Axis.tickCount 10 ] <| yScale y ]
            , g [ transform [ Translate 0 <| h - padding ] ] [ Axis.bottom [ Axis.tickCount 10 ] <| xScale x ]
            , g [] [ drawCurve ( points, x, y ) ]
            ]
        ]


convertPoints : List Float -> List Float -> List ( Float, Float )
convertPoints xs ys =
    List.map2 Tuple.pair xs ys
