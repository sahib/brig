module Modals.Mkdir exposing (Model, Msg, newModel, show, subscriptions, update, view)

import Bootstrap.Alert as Alert
import Bootstrap.Button as Button
import Bootstrap.Form.Input as Input
import Bootstrap.Grid as Grid
import Bootstrap.Grid.Col as Col
import Bootstrap.Modal as Modal
import Browser.Events as Events
import Commands
import Html exposing (..)
import Html.Attributes exposing (..)
import Html.Events exposing (..)
import Http
import Json.Decode as D
import Url
import Util


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
    | KeyPress String String



-- INIT


newModel : Model
newModel =
    { state = Ready
    , modal = Modal.hidden
    , inputName = ""
    , alert = Alert.shown
    }



-- UPDATE


update : Msg -> Model -> ( Model, Cmd Msg )
update msg model =
    case msg of
        CreateDir path ->
            ( model, Commands.doMkdir GotResponse path )

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
            ( { model | modal = Modal.shown, inputName = "" }, Cmd.none )

        ModalClose ->
            ( { model | modal = Modal.hidden, state = Ready }, Cmd.none )

        AlertMsg vis ->
            ( { model | alert = vis }, Cmd.none )

        KeyPress path key ->
            if model.modal == Modal.hidden then
                ( model, Cmd.none )

            else
                case key of
                    "Enter" ->
                        ( model, Commands.doMkdir GotResponse path )

                    _ ->
                        ( model, Cmd.none )



-- VIEW


showPathCollision : Model -> Bool -> Html Msg
showPathCollision model doesExist =
    if doesExist then
        span [ class "text-left" ]
            [ span [ class "fas fa-md fa-exclamation-triangle text-warning" ] []
            , span [ class "text-muted" ]
                [ text (" »" ++ model.inputName ++ "« exists already. Please choose another name.\u{00A0}\u{00A0}\u{00A0}")
                ]
            ]

    else
        span [] []


viewMkdirContent : Model -> List (Grid.Column Msg)
viewMkdirContent model =
    [ Grid.col [ Col.xs12 ]
        [ Input.text
            [ Input.id "mkdir-input"
            , Input.large
            , Input.placeholder "Directory name"
            , Input.onInput InputChanged
            , Input.attrs [ autofocus True ]
            ]
        , br [] []
        , case model.state of
            Ready ->
                text ""

            Fail message ->
                Util.buildAlert model.alert AlertMsg Alert.danger "Oh no!" ("Could not create directory: " ++ message)
        ]
    ]


pathFromUrl : Url.Url -> Model -> String
pathFromUrl url model =
    Util.joinPath [ Util.urlToPath url, model.inputName ]


view : Model -> Url.Url -> (String -> Bool) -> Html Msg
view model url existChecker =
    let
        path =
            Util.urlToPath url

        hasPathCollision =
            existChecker model.inputName
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
            [ showPathCollision model hasPathCollision
            , Button.button
                [ Button.primary
                , Button.attrs
                    [ onClick (CreateDir (pathFromUrl url model))
                    , type_ "submit"
                    , disabled
                        (String.length model.inputName
                            == 0
                            || (case model.state of
                                    Fail _ ->
                                        True

                                    _ ->
                                        False
                               )
                            || hasPathCollision
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


subscriptions : Url.Url -> Model -> Sub Msg
subscriptions url model =
    Sub.batch
        [ Modal.subscriptions model.modal AnimateModal
        , Alert.subscriptions model.alert AlertMsg
        , Events.onKeyPress (D.map (KeyPress <| pathFromUrl url model) <| D.field "key" D.string)
        ]
