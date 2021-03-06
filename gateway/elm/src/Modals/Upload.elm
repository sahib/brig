module Modals.Upload exposing
    ( Model
    , Msg
    , buildButton
    , newModel
    , subscriptions
    , update
    , viewUploadState
    )

import Bootstrap.Alert as Alert
import Bootstrap.Button as Button
import Bootstrap.Grid as Grid
import Bootstrap.Grid.Col as Col
import Bootstrap.Progress as Progress
import Bootstrap.Text as Text
import Commands
import Delay
import Dict
import File
import Html exposing (..)
import Html.Attributes exposing (..)
import Html.Events exposing (..)
import Http
import Json.Decode as D
import List
import Tuple


type alias Alertable =
    { alert : Alert.Visibility
    , path : String
    }


type alias Model =
    { uploads : Dict.Dict String Float
    , failed : List Alertable
    , success : List Alertable
    }


type Msg
    = UploadSelectedFiles String (List File.File)
    | UploadProgress String Http.Progress
    | Uploaded String (Result Http.Error ())
    | UploadCancel String
    | AlertMsg String Alert.Visibility



-- INIT


newModel : Model
newModel =
    { uploads = Dict.empty
    , failed = []
    , success = []
    }


alertMapper : String -> Alert.Visibility -> Alertable -> Alertable
alertMapper path vis a =
    case a.path == path of
        True ->
            { a | alert = vis }

        False ->
            a



-- UPDATE


update : Msg -> Model -> ( Model, Cmd Msg )
update msg model =
    case msg of
        UploadSelectedFiles root files ->
            let
                newUploads =
                    Dict.union model.uploads <| Dict.fromList (List.map (\f -> ( File.name f, 0 )) files)
            in
            ( { model | uploads = newUploads }
            , Cmd.batch (List.map (Commands.doUpload Uploaded root) files)
            )

        UploadProgress path progress ->
            case progress of
                Http.Sending p ->
                    ( { model | uploads = Dict.insert path (Http.fractionSent p) model.uploads }, Cmd.none )

                Http.Receiving _ ->
                    ( model, Cmd.none )

        Uploaded path result ->
            let
                newUploads =
                    Dict.remove path model.uploads
            in
            case result of
                Ok _ ->
                    ( { model
                        | uploads = newUploads
                        , success = Alertable Alert.shown path :: model.success
                      }
                    , Delay.after 5 Delay.Second (AlertMsg path Alert.closed)
                    )

                Err _ ->
                    ( { model
                        | uploads = newUploads
                        , failed = Alertable Alert.shown path :: model.failed
                      }
                    , Delay.after 30 Delay.Second (AlertMsg path Alert.closed)
                    )

        UploadCancel path ->
            ( { model | uploads = Dict.remove path model.uploads }
            , Http.cancel ("upload-" ++ path)
            )

        AlertMsg path vis ->
            ( { model
                | success = List.map (alertMapper path vis) model.success
                , failed = List.map (alertMapper path vis) model.failed
              }
            , Cmd.none
            )



-- VIEW


filesDecoder : D.Decoder (List File.File)
filesDecoder =
    D.at [ "target", "files" ] (D.list File.decoder)


buildButton : Model -> Bool -> String -> (Msg -> msg) -> Html msg
buildButton model currIsFile currRoot toMsg =
    label
        [ class "btn btn-file btn-link btn-default text-left"
        , id "action-btn"
        , if currIsFile then
            class "disabled"

          else
            class "btn-default"
        ]
        [ span [ class "fas fa-plus" ] []
        , span [ class "d-lg-inline d-none" ] [ text "\u{00A0}\u{00A0}Upload" ]
        , input
            [ type_ "file"
            , multiple True
            , on "change"
                (D.map toMsg
                    (D.map
                        (UploadSelectedFiles currRoot)
                        filesDecoder
                    )
                )
            , style "display" "none"
            , disabled currIsFile
            ]
            []
        ]


clampText : String -> Int -> String
clampText text length =
    if String.length text <= length then
        text

    else
        String.slice 0 length text ++ "…"


viewAlert : Alert.Visibility -> String -> Bool -> Html Msg
viewAlert alert path isSuccess =
    Alert.config
        |> Alert.dismissableWithAnimation (AlertMsg path)
        |> (if isSuccess then
                Alert.success

            else
                Alert.danger
           )
        |> Alert.children
            [ Grid.row []
                [ Grid.col [ Col.xs10 ]
                    [ span
                        [ if isSuccess then
                            class "fas fa-xs fa-check"

                          else
                            class "fas fa-xs fa-exclamation-circle"
                        ]
                        []
                    , text (" " ++ clampText path 15)
                    ]
                , Grid.col [ Col.xs2, Col.textAlign Text.alignXsRight ]
                    [ Button.button
                        [ Button.roleLink
                        , Button.attrs
                            [ class "notification-close-btn"
                            , onClick (AlertMsg path Alert.closed)
                            ]
                        ]
                        [ span [ class "fas fa-xs fa-times" ] [] ]
                    ]
                ]
            ]
        |> Alert.view alert


viewProgressIndicator : String -> Float -> Html Msg
viewProgressIndicator path fraction =
    Grid.row []
        [ Grid.col [ Col.md10 ]
            [ Progress.progress
                [ Progress.value (100 * fraction)
                , Progress.customLabel [ text (clampText path 25) ]
                , Progress.attrs [ style "height" "25px" ]
                , Progress.wrapperAttrs [ style "height" "25px" ]
                ]
            ]
        , Grid.col [ Col.md2 ]
            [ Button.button
                [ Button.roleLink
                , Button.attrs [ class "progress-cancel", onClick (UploadCancel path) ]
                ]
                [ span [ class "fas fa-xs fa-times" ] [] ]
            ]
        ]


viewUploadState : Model -> Html Msg
viewUploadState model =
    div []
        [ br [] []
        , br [] []
        , ul [ class "notification-list list-group" ]
            (List.map (\a -> viewAlert a.alert a.path True) model.success
                ++ List.map (\a -> viewAlert a.alert a.path False) model.failed
                ++ List.map (\p -> viewProgressIndicator (Tuple.first p) (Tuple.second p)) (Dict.toList model.uploads)
            )
        ]



-- SUBSCRIPTIONS


subscriptions : Model -> Sub Msg
subscriptions model =
    Sub.batch
        [ Sub.batch
            (List.map
                (\p -> Http.track ("upload-" ++ p) (UploadProgress p))
                (Dict.keys model.uploads)
            )
        , Sub.batch (List.map (\a -> Alert.subscriptions a.alert (AlertMsg a.path)) model.success)
        , Sub.batch (List.map (\a -> Alert.subscriptions a.alert (AlertMsg a.path)) model.failed)
        ]
