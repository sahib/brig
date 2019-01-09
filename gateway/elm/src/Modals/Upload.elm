module Modals.Upload exposing (Model, Msg, newModel, show, subscriptions, update, view)

import Bootstrap.Button as Button
import Bootstrap.Grid as Grid
import Bootstrap.Grid.Col as Col
import Bootstrap.Grid.Row as Row
import Bootstrap.Modal as Modal
import Bootstrap.Progress as Progress
import Browser
import Browser.Navigation as Nav
import File
import Html exposing (..)
import Html.Attributes exposing (..)
import Html.Events exposing (..)
import Http
import Json.Decode as D
import List
import Url



-- TODO: Handle case where the file already exist.
--       Warn and ask to overwrite?


type State
    = Ready
    | Uploading Float
    | Done
    | Fail


type alias Model =
    { state : State
    , modal : Modal.Visibility
    }


type Msg
    = UploadSelectedFiles (List File.File)
    | UploadProgress Http.Progress
    | Uploaded (Result Http.Error ())
    | UploadCancel
    | ModalShow
    | AnimateModal Modal.Visibility
    | ModalClose



-- INIT


newModel : Model
newModel =
    { state = Ready, modal = Modal.hidden }



-- UPDATE


update : Msg -> Model -> ( Model, Cmd Msg )
update msg model =
    case msg of
        UploadSelectedFiles files ->
            ( { model | state = Uploading 0 }
            , Http.request
                { method = "POST"
                , url = "/api/v0/upload"
                , headers = []
                , body = Http.multipartBody <| List.map (Http.filePart "files[]") files
                , expect = Http.expectWhatever Uploaded
                , timeout = Nothing
                , tracker = Just "upload"
                }
            )

        UploadProgress progress ->
            case progress of
                Http.Sending p ->
                    ( { model | state = Uploading <| Http.fractionSent p }, Cmd.none )

                Http.Receiving _ ->
                    ( model, Cmd.none )

        Uploaded result ->
            case result of
                Ok _ ->
                    ( { model | state = Done }, Cmd.none )

                Err _ ->
                    ( { model | state = Fail }, Cmd.none )

        UploadCancel ->
            ( { model | state = Ready }, Cmd.none )

        AnimateModal visibility ->
            ( { model | modal = visibility }, Cmd.none )

        ModalShow ->
            ( { model | modal = Modal.shown }, Cmd.none )

        ModalClose ->
            -- Clear up failure or success state on close.
            let
                updatedModel =
                    case model.state of
                        -- TODO: Is there a way to pattern match Done and Fail?
                        Done ->
                            { model | state = Ready }

                        Fail ->
                            { model | state = Ready }

                        _ ->
                            model
            in
            ( { updatedModel | modal = Modal.hidden }, Cmd.none )



-- VIEW


filesDecoder : D.Decoder (List File.File)
filesDecoder =
    D.at [ "target", "files" ] (D.list File.decoder)


viewUploadState : Model -> List (Grid.Column Msg)
viewUploadState model =
    case model.state of
        Ready ->
            [ Grid.col [ Col.xs12 ]
                [ label [ class "btn btn-file btn-primary btn-default" ]
                    [ text "Browse local files"
                    , input
                        [ type_ "file"
                        , multiple True
                        , on "change" (D.map UploadSelectedFiles filesDecoder)
                        , style "display" "none"
                        ]
                        []
                    ]
                ]
            ]

        Uploading fraction ->
            [ Grid.col [ Col.xs10 ]
                [ Progress.progress
                    [ Progress.animated
                    , Progress.value (100 * fraction)
                    ]
                ]
            , Grid.col [ Col.xs2 ]
                [ Button.button
                    [ Button.outlinePrimary
                    , Button.attrs [ onClick UploadCancel ]
                    ]
                    [ text "Cancel" ]
                ]
            ]

        Done ->
            [ Grid.col [ Col.xs12 ] [ text "Upload done" ] ]

        Fail ->
            [ Grid.col [ Col.xs12 ] [ text "Upload failed for unknown reasons" ] ]


view : Model -> Html Msg
view model =
    Modal.config ModalClose
        |> Modal.large
        |> Modal.withAnimation AnimateModal
        |> Modal.h5 [] [ text "Upload a new file" ]
        |> Modal.body []
            [ Grid.containerFluid []
                [ Grid.row [] (viewUploadState model) ]
            ]
        |> Modal.footer []
            [ Button.button
                [ Button.outlinePrimary
                , Button.attrs [ onClick <| AnimateModal Modal.hiddenAnimated ]
                ]
                [ text "Close" ]
            ]
        |> Modal.view model.modal


show : Msg
show =
    ModalShow



-- SUBSCRIPTIONS


subscriptions : Model -> Sub Msg
subscriptions model =
    Sub.batch
        [ Http.track "upload" UploadProgress
        , Modal.subscriptions model.modal AnimateModal
        ]
