module Modals.RemoteFolders exposing
    ( Model
    , Msg
    , newModel
    , show
    , subscriptions
    , update
    , view
    )

import Bootstrap.Alert as Alert
import Bootstrap.Button as Button
import Bootstrap.Form.Input as Input
import Bootstrap.Form.InputGroup as InputGroup
import Bootstrap.Grid as Grid
import Bootstrap.Grid.Col as Col
import Bootstrap.Grid.Row as Row
import Bootstrap.ListGroup as ListGroup
import Bootstrap.Modal as Modal
import Bootstrap.Text as Text
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
    , folders : List String
    , newFolder : String
    , remote : Commands.Remote
    , modal : Modal.Visibility
    , alert : Alert.Visibility
    }


type Msg
    = RemoteModify
    | ModalShow Commands.Remote
    | NewFolderChange String
    | FolderRemove String
    | FolderAdd
    | GotResponse (Result Http.Error String)
    | AnimateModal Modal.Visibility
    | AlertMsg Alert.Visibility
    | ModalClose
    | KeyPress String



-- INIT


newModel : Model
newModel =
    newModelWithState Modal.hidden Commands.emptyRemote


newModelWithState : Modal.Visibility -> Commands.Remote -> Model
newModelWithState state remote =
    { state = Ready
    , modal = state
    , folders = []
    , newFolder = ""
    , remote = remote
    , alert = Alert.shown
    }



-- UPDATE


submit : Model -> Cmd Msg
submit model =
    Commands.doRemoteModify GotResponse model.remote


fixFolder : String -> String
fixFolder path =
    Util.prefixSlash path


addFolder : Model -> ( Model, Cmd Msg )
addFolder model =
    let
        oldRemote =
            model.remote

        cleanFolder =
            fixFolder model.newFolder

        newRemote =
            { oldRemote | folders = List.sort <| cleanFolder :: oldRemote.folders }
    in
    ( { model | remote = newRemote, newFolder = "" }, Cmd.none )


update : Msg -> Model -> ( Model, Cmd Msg )
update msg model =
    case msg of
        GotResponse result ->
            case result of
                Ok _ ->
                    -- New list model means also new checked entries.
                    ( { model | state = Ready, modal = Modal.hidden }, Cmd.none )

                Err err ->
                    ( { model | state = Fail <| Util.httpErrorToString err }, Cmd.none )

        NewFolderChange folder ->
            ( { model | newFolder = folder }, Cmd.none )

        FolderAdd ->
            addFolder model

        FolderRemove folder ->
            let
                oldRemote =
                    model.remote

                newRemote =
                    { oldRemote | folders = List.filter (\f -> f /= folder) oldRemote.folders }
            in
            ( { model | remote = newRemote }, Cmd.none )

        RemoteModify ->
            ( model, submit model )

        AnimateModal visibility ->
            ( { model | modal = visibility }, Cmd.none )

        ModalShow remote ->
            ( newModelWithState Modal.shown remote, Cmd.none )

        ModalClose ->
            ( { model | modal = Modal.hidden }, Cmd.none )

        AlertMsg vis ->
            ( { model | alert = vis }, Cmd.none )

        KeyPress key ->
            if model.modal == Modal.hidden then
                ( model, Cmd.none )

            else
                case key of
                    "Enter" ->
                        addFolder model

                    _ ->
                        ( model, Cmd.none )



-- VIEW


viewRow : Html Msg -> Html Msg -> Html Msg -> Html Msg
viewRow a b c =
    Grid.row []
        [ Grid.col [ Col.xs1, Col.textAlign Text.alignXsRight ] [ a ]
        , Grid.col [ Col.xs10, Col.textAlign Text.alignXsLeft ] [ b ]
        , Grid.col [ Col.xs1, Col.textAlign Text.alignXsLeft ] [ c ]
        ]


viewFolder : String -> Html Msg
viewFolder folder =
    viewRow
        (span [ class "fas fa-md fa-folder text-muted" ] [])
        (text folder)
        (Button.button
            [ Button.attrs [ class "close", onClick <| FolderRemove folder ] ]
            [ span [ class "fas fa-xs fa-times text-muted" ] []
            ]
        )


viewFolders : Commands.Remote -> Html Msg
viewFolders remote =
    if List.length remote.folders <= 0 then
        span
            [ class "text-muted text-center" ]
            [ text "No folders. This means this user can see everthing."
            , br [] []
            , text "Add a new folder below to change this."
            , br [] []
            , br [] []
            ]

    else
        ListGroup.ul
            (List.map (\f -> ListGroup.li [] [ viewFolder f ]) remote.folders)


viewNewFolderControl : Model -> Html Msg
viewNewFolderControl model =
    InputGroup.config
        (InputGroup.text
            [ Input.placeholder "New folder path"
            , Input.onInput NewFolderChange
            , Input.value model.newFolder
            ]
        )
        |> InputGroup.predecessors
            [ InputGroup.button
                [ Button.primary
                , Button.attrs
                    [ onClick <| FolderAdd
                    , disabled
                        (String.length model.newFolder == 0)
                    ]
                ]
                [ span [ class "fas fa-lg fa-folder-plus" ] [] ]
            ]
        |> InputGroup.view


viewRemoteAddContent : Model -> List (Grid.Column Msg)
viewRemoteAddContent model =
    [ Grid.col [ Col.xs12 ]
        [ viewFolders model.remote
        , br [] []
        , viewNewFolderControl model
        , case model.state of
            Ready ->
                text ""

            Fail message ->
                Util.buildAlert model.alert AlertMsg Alert.danger "Oh no!" ("Could not add remote: " ++ message)
        ]
    ]


view : Model -> Html Msg
view model =
    Modal.config ModalClose
        |> Modal.large
        |> Modal.withAnimation AnimateModal
        |> Modal.header [ class "modal-title modal-header-primary" ]
            [ h4 [] [ text "Edit folders" ] ]
        |> Modal.body []
            [ Grid.containerFluid []
                [ Grid.row
                    [ Row.attrs [ style "min-width" "60vh" ]
                    ]
                    (viewRemoteAddContent model)
                ]
            ]
        |> Modal.footer []
            [ Button.button
                [ Button.primary
                , Button.attrs
                    [ onClick RemoteModify
                    , type_ "submit"
                    , disabled False
                    ]
                ]
                [ text "Save" ]
            , Button.button
                [ Button.outlinePrimary
                , Button.attrs [ onClick <| AnimateModal Modal.hiddenAnimated ]
                ]
                [ text "Cancel" ]
            ]
        |> Modal.view model.modal


show : Commands.Remote -> Msg
show remote =
    ModalShow remote



-- SUBSCRIPTIONS


subscriptions : Model -> Sub Msg
subscriptions model =
    Sub.batch
        [ Modal.subscriptions model.modal AnimateModal
        , Alert.subscriptions model.alert AlertMsg
        , Events.onKeyPress (D.map KeyPress <| D.field "key" D.string)
        ]
