module Modals.Mkdir exposing (Model, Msg, newModel, show, subscriptions, update, view)

import Bootstrap.Alert as Alert
import Bootstrap.Button as Button
import Bootstrap.Form.Input as Input
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
import Json.Encode as E
import List
import Url
import Util



-- TODO: Handle case where the dir already exist.
--       Warn and ask to overwrite?


type State
    = Ready
    | Fail String


type alias Model =
    { state : State
    , inputName : String
    , modal : Modal.Visibility
    , alert : Alert.Visibility
    }


type Msg
    = CreateDir String
    | InputChanged String
    | ModalShow
    | GotResponse (Result Http.Error String)
    | AnimateModal Modal.Visibility
    | AlertMsg Alert.Visibility
    | ModalClose



-- INIT


newModel : Model
newModel =
    { state = Ready
    , modal = Modal.hidden
    , inputName = ""
    , alert = Alert.shown
    }



-- UPDATE


type alias Query =
    { path : String
    }


encode : Query -> E.Value
encode q =
    E.object
        [ ( "path", E.string q.path ) ]


decode : D.Decoder String
decode =
    D.field "message" D.string


update : Msg -> Model -> ( Model, Cmd Msg )
update msg model =
    case msg of
        CreateDir path ->
            ( model
            , Http.post
                { url = "/api/v0/mkdir"
                , body = Http.jsonBody <| encode <| Query path
                , expect = Http.expectJson GotResponse decode
                }
            )

        InputChanged inputName ->
            ( { model | inputName = inputName }, Cmd.none )

        GotResponse result ->
            case result of
                Ok _ ->
                    -- New list model means also new checked entries.
                    ( { model | state = Ready, modal = Modal.hidden }, Cmd.none )

                Err err ->
                    ( { model | state = Fail <| Util.httpErrorToString err }, Cmd.none )

        AnimateModal visibility ->
            ( { model | modal = visibility }, Cmd.none )

        ModalShow ->
            ( { model | modal = Modal.shown }, Cmd.none )

        ModalClose ->
            ( { model | modal = Modal.hidden, state = Ready }, Cmd.none )

        AlertMsg vis ->
            ( { model | alert = vis }, Cmd.none )



-- VIEW


viewMkdirContent : Model -> List (Grid.Column Msg)
viewMkdirContent model =
    [ Grid.col [ Col.xs12 ]
        [ Input.text
            [ Input.id "mkdir-input"
            , Input.large
            , Input.placeholder "Directory name"
            , Input.onInput InputChanged
            ]
        , br [] []
        , case model.state of
            Ready ->
                text ""

            Fail message ->
                Util.buildAlert model.alert AlertMsg True "Oh no!" ("Could not create directory: " ++ message)
        ]
    ]


view : Model -> Url.Url -> Html Msg
view model url =
    let
        path =
            Util.urlToPath url
    in
    Modal.config ModalClose
        |> Modal.large
        |> Modal.withAnimation AnimateModal
        |> Modal.h5 [] [ text ("Create a new directory in " ++ path) ]
        |> Modal.body []
            [ Grid.containerFluid []
                [ Grid.row [] (viewMkdirContent model) ]
            ]
        |> Modal.footer []
            [ Button.button
                [ Button.primary
                , Button.attrs
                    [ onClick <| CreateDir (Util.joinPath [ path, model.inputName ])
                    , disabled
                        (String.length model.inputName
                            == 0
                            || (case model.state of
                                    Fail _ ->
                                        True

                                    _ ->
                                        False
                               )
                        )
                    ]
                ]
                [ text "Create" ]
            , Button.button
                [ Button.outlinePrimary
                , Button.attrs [ onClick <| AnimateModal Modal.hiddenAnimated ]
                ]
                [ text "Cancel" ]
            ]
        |> Modal.view model.modal


show : Msg
show =
    ModalShow



-- SUBSCRIPTIONS


subscriptions : Model -> Sub Msg
subscriptions model =
    Sub.batch
        [ Modal.subscriptions model.modal AnimateModal
        , Alert.subscriptions model.alert AlertMsg
        ]
