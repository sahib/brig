module Modals.Remove exposing (Model, Msg, newModel, show, subscriptions, update, view)

import Bootstrap.Alert as Alert
import Bootstrap.Button as Button
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


import Ls

import Util


type State
    = Ready
    | Fail String


type alias Model =
    { state : State
    , modal : Modal.Visibility
    , alert : Alert.Visibility
    }


type Msg
    = RemoveAll (List String)
    | ModalShow
    | GotResponse (Result Http.Error String)
    | AnimateModal Modal.Visibility
    | AlertMsg Alert.Visibility
    | ModalClose
    | KeyPress (List String) String



-- INIT


newModel : Model
newModel =
    { state = Ready
    , modal = Modal.hidden
    , alert = Alert.shown
    }



-- UPDATE


update : Msg -> Model -> ( Model, Cmd Msg )
update msg model =
    case msg of
        RemoveAll paths ->
            ( model, Commands.doRemove GotResponse paths )

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

        KeyPress paths key ->
            ( model
            , if model.modal == Modal.hidden then
                Cmd.none

              else
                case key of
                    "Enter" ->
                        Commands.doRemove GotResponse paths

                    _ ->
                        Cmd.none
            )



-- VIEW


viewRemoveContent : Model -> Ls.Model -> List (Grid.Column Msg)
viewRemoveContent model lsModel =
    [ Grid.col [ Col.xs12 ]
        [ case model.state of
            Ready ->
                text ("Remove the " ++ String.fromInt (Ls.nSelectedItems lsModel) ++ " selected items")

            Fail message ->
                Util.buildAlert model.alert AlertMsg Alert.danger "Oh no!" ("Could not remove directory: " ++ message)
        ]
    ]


view : Model -> Ls.Model -> Html Msg
view model lsModel =
    Modal.config ModalClose
        |> Modal.large
        |> Modal.withAnimation AnimateModal
        |> Modal.h5 [] [ text "Really remove?" ]
        |> Modal.body []
            [ Grid.containerFluid []
                [ Grid.row [] (viewRemoveContent model lsModel) ]
            ]
        |> Modal.footer []
            [ Button.button
                [ Button.danger
                , Button.attrs
                    [ onClick <| RemoveAll <| Ls.selectedPaths lsModel
                    , disabled
                        (case model.state of
                            Fail _ ->
                                True

                            _ ->
                                False
                        )
                    ]
                ]
                [ text "Remove" ]
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


subscriptions : Ls.Model -> Model -> Sub Msg
subscriptions lsModel model =
    Sub.batch
        [ Modal.subscriptions model.modal AnimateModal
        , Alert.subscriptions model.alert AlertMsg
        , Events.onKeyPress (D.map (KeyPress <| Ls.selectedPaths lsModel) <| D.field "key" D.string)
        ]
