module Modals.Upload exposing
    ( Model
    , Msg
    , buildButton
    , newModel
    , show
    , subscriptions
    , update
    , viewUploadState
    )

import Bootstrap.Alert as Alert
import Bootstrap.Button as Button
import Bootstrap.Grid as Grid
import Bootstrap.Grid.Col as Col
import Bootstrap.Grid.Row as Row
import Bootstrap.Modal as Modal
import Bootstrap.Progress as Progress
import Bootstrap.Text as Text
import Browser
import Browser.Navigation as Nav
import Delay
import Dict
import File
import Html exposing (..)
import Html.Attributes exposing (..)
import Html.Events exposing (..)
import Http
import Json.Decode as D
import List
import Ls
import Task
import Tuple
import Url
import Util



-- TODO: Make "Upload" button open the file dialog directly.


type alias Alertable =
    { alert : Alert.Visibility
    , path : String
    }


type alias Model =
    { uploads : Dict.Dict String Float
    , failed : List Alertable
    , success : List Alertable
    , modal : Modal.Visibility
    }


type Msg
    = UploadSelectedFiles String (List File.File)
    | UploadProgress String Http.Progress
    | Uploaded String (Result Http.Error ())
    | UploadCancel String
    | ModalShow
    | AnimateModal Modal.Visibility
    | ModalClose
    | AlertMsg String Alert.Visibility



-- INIT


newModel : Model
newModel =
    { uploads = Dict.empty
    , modal = Modal.hidden
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
            ( { model | uploads = newUploads, modal = Modal.hidden }
            , Cmd.batch
                (List.map
                    (\f ->
                        Http.request
                            { method = "POST"
                            , url = "/api/v0/upload?root=" ++ Url.percentEncode root
                            , headers = []
                            , body = Http.multipartBody [ Http.filePart "files[]" f ]
                            , expect = Http.expectWhatever (Uploaded (File.name f))
                            , timeout = Nothing
                            , tracker = Just ("upload-" ++ File.name f)
                            }
                    )
                    files
                )
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

        AnimateModal visibility ->
            ( { model | modal = visibility }, Cmd.none )

        ModalShow ->
            ( { model | modal = Modal.shown }, Cmd.none )

        ModalClose ->
            ( { model | modal = Modal.hidden }, Cmd.none )

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


show : Msg
show =
    ModalShow


buildButton : Model -> Ls.Model -> (Msg -> msg) -> Html msg
buildButton model lsModel toMsg =
    let
        isDisabled =
            Ls.currIsFile lsModel
    in
    label
        [ class "btn btn-file btn-link btn-default"
        , id "upload-btn"
        , if isDisabled then
            class "disabled"

          else
            class "btn-default"
        ]
        [ span [ class "fas fa-plus" ] []
        , text "\u{00A0}\u{00A0}Upload"
        , input
            [ type_ "file"
            , multiple True
            , on "change"
                (D.map toMsg
                    (D.map
                        (UploadSelectedFiles
                            (Maybe.withDefault "/" (Ls.currRoot lsModel))
                        )
                        filesDecoder
                    )
                )
            , style "display" "none"
            , disabled isDisabled
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
                    , text (" " ++ clampText path 25)
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
        , Modal.subscriptions model.modal AnimateModal
        ]
