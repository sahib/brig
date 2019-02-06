module Util exposing
    ( AlertState
    , AlertType(..)
    , basename
    , buildAlert
    , defaultAlertState
    , dirname
    , formatLastModified
    , formatLastModifiedOwner
    , httpErrorToString
    , joinPath
    , monthToInt
    , prefixSlash
    , splitPath
    , urlEncodePath
    , urlPrefixToString
    , urlToPath
    , viewAlert
    , viewToggleSwitch
    )

import Bootstrap.Alert as Alert
import Bootstrap.Button as Button
import Bootstrap.Grid as Grid
import Bootstrap.Grid.Col as Col
import Bootstrap.Grid.Row as Row
import Bootstrap.Text as Text
import Html exposing (..)
import Html.Attributes exposing (..)
import Html.Events exposing (..)
import Http
import Time
import Url


boolToStr : Bool -> String
boolToStr b =
    case b of
        True ->
            "yes"

        False ->
            "no"


monthToInt : Time.Month -> Int
monthToInt month =
    -- This feels stupid.
    case month of
        Time.Jan ->
            1

        Time.Feb ->
            2

        Time.Mar ->
            3

        Time.Apr ->
            4

        Time.May ->
            5

        Time.Jun ->
            6

        Time.Jul ->
            7

        Time.Aug ->
            8

        Time.Sep ->
            9

        Time.Oct ->
            10

        Time.Nov ->
            11

        Time.Dec ->
            12


formatLastModifiedOwner : Time.Zone -> Time.Posix -> String -> Html.Html msg
formatLastModifiedOwner z t owner =
    p [] [ text (formatLastModified z t), span [ class "text-muted" ] [ text " by " ], text owner ]


formatLastModified : Time.Zone -> Time.Posix -> String
formatLastModified z t =
    String.join " "
        -- Day portion:
        [ String.join
            "/"
            [ Time.toDay z t |> String.fromInt
            , Time.toMonth z t |> monthToInt |> String.fromInt
            , Time.toYear z t |> String.fromInt
            ]

        -- Time portion:
        , String.join ":"
            [ Time.toHour z t |> String.fromInt |> String.padLeft 2 '0'
            , Time.toMinute z t |> String.fromInt |> String.padLeft 2 '0'
            , Time.toSecond z t |> String.fromInt |> String.padLeft 2 '0'
            ]
        ]


splitPath : String -> List String
splitPath path =
    List.filter (\s -> String.length s > 0) (String.split "/" path)


joinPath : List String -> String
joinPath paths =
    "/" ++ String.join "/" (List.foldr (++) [] (List.map splitPath paths))


urlToPath : Url.Url -> String
urlToPath url =
    let
        decodeUrlPart =
            \e ->
                case Url.percentDecode e of
                    Just val ->
                        val

                    Nothing ->
                        ""
    in
    case splitPath url.path of
        [] ->
            "/"

        _ :: xs ->
            "/" ++ String.join "/" (List.map decodeUrlPart xs)


basename : String -> String
basename path =
    let
        split =
            List.reverse (splitPath path)
    in
    case split of
        [] ->
            "/"

        x :: _ ->
            x


prefixSlash : String -> String
prefixSlash path =
    if String.startsWith "/" path then
        path

    else
        "/" ++ path


dirname : String -> String
dirname path =
    let
        split =
            splitPath path
    in
    case split of
        [] ->
            "/"

        _ ->
            joinPath <| List.take (List.length split - 1) split


buildAlert : Alert.Visibility -> (Alert.Visibility -> msg) -> (Alert.Config msg -> Alert.Config msg) -> String -> String -> Html msg
buildAlert visibility msg severity title message =
    Alert.config
        |> Alert.dismissableWithAnimation msg
        |> severity
        |> Alert.children
            [ if String.length title > 0 then
                Alert.h4 [] [ text title ]

              else
                text ""
            , text message
            ]
        |> Alert.view visibility


httpErrorToString : Http.Error -> String
httpErrorToString err =
    case err of
        Http.BadUrl msg ->
            "Bad url: " ++ msg

        Http.Timeout ->
            "Timeout"

        Http.NetworkError ->
            "Network error"

        Http.BadStatus status ->
            "Bad status: " ++ String.fromInt status

        Http.BadBody msg ->
            "Could not decode body: " ++ msg


urlPrefixToString : Url.Url -> String
urlPrefixToString url =
    (case url.protocol of
        Url.Https ->
            "https://"

        Url.Http ->
            "http://"
    )
        ++ url.host
        ++ (case url.port_ of
                Just port_ ->
                    ":" ++ String.fromInt port_

                Nothing ->
                    ""
           )
        ++ "/"


urlEncodePath : String -> String
urlEncodePath path =
    joinPath (List.map Url.percentEncode (splitPath path))


viewToggleSwitch : (Bool -> msg) -> String -> Bool -> Html msg
viewToggleSwitch toMsg message isChecked =
    span []
        [ span []
            [ label [ class "toggle-switch" ]
                [ input [ type_ "checkbox", onCheck toMsg, checked isChecked ] []
                , span [ class "toggle-slider toggle-round" ] []
                ]
            ]
        , span
            [ class "text-muted" ]
            [ text (" " ++ message) ]
        ]



-- ALERT UTILS


type AlertType
    = Danger
    | Success
    | Info


type alias AlertState =
    { message : String
    , typ : AlertType
    , vis : Alert.Visibility
    }


defaultAlertState : AlertState
defaultAlertState =
    { message = ""
    , typ = Info
    , vis = Alert.closed
    }


iconFromAlertType : AlertType -> Html msg
iconFromAlertType typ =
    case typ of
        Danger ->
            span [ class "fas fa-xs fa-times" ] []

        Success ->
            span [ class "fas fa-xs fa-check" ] []

        _ ->
            text ""


visualFromAlertType : AlertType -> (Alert.Config msg -> Alert.Config msg)
visualFromAlertType typ =
    case typ of
        Danger ->
            Alert.danger

        Success ->
            Alert.success

        _ ->
            Alert.info


viewAlert : (Alert.Visibility -> msg) -> AlertState -> Html msg
viewAlert toMsg alert =
    Alert.config
        |> Alert.dismissableWithAnimation toMsg
        |> visualFromAlertType alert.typ
        |> Alert.children
            [ Grid.row []
                [ Grid.col [ Col.xs10 ]
                    [ iconFromAlertType alert.typ
                    , text (" " ++ alert.message)
                    ]
                , Grid.col [ Col.xs2, Col.textAlign Text.alignXsRight ]
                    [ Button.button
                        [ Button.roleLink
                        , Button.attrs
                            [ class "notification-close-btn"
                            , onClick (toMsg Alert.closed)
                            ]
                        ]
                        [ span [ class "fas fa-xs fa-times" ] [] ]
                    ]
                ]
            ]
        |> Alert.view alert.vis
