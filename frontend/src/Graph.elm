module Graph exposing (convertPoints, viewGraph)

import Axis
import Color exposing (Color)
import Html
import Html.Attributes exposing (class, list)
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


drawCurve : ( Float, Float ) -> ( Float, Float ) -> ( List ( Float, Float ), Color ) -> Html.Html msg
drawCurve xBounds yBounds ( data, clr ) =
    List.map Just (preparedPoints (xScale xBounds) (yScale yBounds) data)
        |> Shape.line Shape.linearCurve
        |> (\path -> Path.element path [ stroke (Paint clr), fill PaintNone, strokeWidth 1 ])


viewGraph : List ( List ( Float, Float ), Color ) -> ( Float, Float ) -> Html.Html msg
viewGraph points_data bounds =
    let
        y_bounds =
            points_data
                |> List.map Tuple.first
                |> getOnRange bounds
                |> getMinMax
    in
    Html.div []
        [ svg [ viewBox 0 0 w h ]
            [ g [ transform [ Translate padding 0 ] ] [ Axis.left [ Axis.tickCount 10 ] <| yScale y_bounds ]
            , g [ transform [ Translate 0 <| h - padding ] ] [ Axis.bottom [ Axis.tickCount 10 ] <| xScale bounds ]
            , g [] <| List.map (\data -> drawCurve bounds y_bounds data) points_data
            ]
        ]


convertPoints : List Float -> List Float -> List ( Float, Float )
convertPoints xs ys =
    List.map2 Tuple.pair xs ys


getMinMax : List (List ( Float, Float )) -> ( Float, Float )
getMinMax list =
    let
        max l =
            case List.maximum (List.map Tuple.second l) of
                Just val ->
                    val + 0.1

                Nothing ->
                    1

        min l =
            case List.minimum (List.map Tuple.second l) of
                Just val ->
                    val

                Nothing ->
                    0

        mins =
            case List.minimum (List.map min list) of
                Just val ->
                    val

                Nothing ->
                    0

        maxs =
            case List.maximum (List.map max list) of
                Just val ->
                    val + 0.01

                Nothing ->
                    9
    in
    ( mins, maxs )


getOnRange : ( Float, Float ) -> List (List ( Float, Float )) -> List (List ( Float, Float ))
getOnRange boundries lists =
    let
        filt : List ( Float, Float ) -> List ( Float, Float )
        filt list =
            List.filter (\x -> Tuple.first x >= Tuple.first boundries && Tuple.first x <= Tuple.second boundries) list
    in
    List.map filt lists
