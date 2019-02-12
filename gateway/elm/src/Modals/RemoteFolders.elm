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
import Modals.MoveCopy as MoveCopy
import Set
import Util


type State
    = Ready
    | Fail String


type alias Model =
    { state : State
    , allDirs : List String
    , filter : String
    , remote : Commands.Remote
    , modal : Modal.Visibility
    , alert : Alert.Visibility
    }


type Msg
    = ModalShow Commands.Remote
    | FolderRemove String
    | GotResponse (Result Http.Error String)
    | AnimateModal Modal.Visibility
    | AlertMsg Alert.Visibility
    | ModalClose
    | GotAllDirsResponse (Result Http.Error (List String))
    | DirChosen String
    | SearchInput String



-- INIT


newModel : Model
newModel =
    newModelWithState Modal.hidden Commands.emptyRemote


newModelWithState : Modal.Visibility -> Commands.Remote -> Model
newModelWithState state remote =
    { state = Ready
    , modal = state
    , allDirs = []
    , filter = ""
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


addFolder : Model -> String -> ( Model, Cmd Msg )
addFolder model folder =
    let
        oldRemote =
            model.remote

        cleanFolder =
            fixFolder folder

        newRemote =
            { oldRemote | folders = List.sort <| cleanFolder :: oldRemote.folders }

        upModel =
            { model | remote = newRemote }
    in
    ( upModel, submit upModel )


update : Msg -> Model -> ( Model, Cmd Msg )
update msg model =
    case msg of
        GotResponse result ->
            case result of
                Ok _ ->
                    -- New list model means also new checked entries.
                    ( { model | state = Ready }, Cmd.none )

                Err err ->
                    ( { model | state = Fail <| Util.httpErrorToString err }, Cmd.none )

        FolderRemove folder ->
            let
                oldRemote =
                    model.remote

                newRemote =
                    { oldRemote | folders = List.filter (\f -> f /= folder) oldRemote.folders }

                upModel =
                    { model | remote = newRemote }
            in
            ( upModel, submit upModel )

        AnimateModal visibility ->
            ( { model | modal = visibility }, Cmd.none )

        ModalShow remote ->
            ( newModelWithState Modal.shown remote
            , Commands.doListAllDirs GotAllDirsResponse
            )

        GotAllDirsResponse result ->
            case result of
                Ok allDirs ->
                    ( { model | allDirs = allDirs }, Cmd.none )

                Err _ ->
                    ( model, Cmd.none )

        DirChosen choice ->
            addFolder model choice

        SearchInput filter ->
            ( { model | filter = filter }, Cmd.none )

        ModalClose ->
            ( { model | modal = Modal.hidden, filter = "" }, Cmd.none )

        AlertMsg vis ->
            ( { model | alert = vis }, Cmd.none )



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
    let
        folders =
            List.sort (Set.toList (Set.fromList remote.folders))
    in
    if List.length folders <= 0 then
        span
            [ class "text-muted text-center" ]
            [ text "No folders. This means this user can see everthing."
            , br [] []
            , text "Add a new folder below to limit what this remote can see."
            , br [] []
            , br [] []
            ]

    else
        ListGroup.ul
            (List.map (\f -> ListGroup.li [] [ viewFolder f ]) folders)


viewRemoteFoldersContent : Model -> List (Grid.Column Msg)
viewRemoteFoldersContent model =
    [ Grid.col [ Col.xs12 ]
        [ h4 [] [ span [ class "text-muted text-center" ] [ text "Visible folders" ] ]
        , viewFolders model.remote
        , br [] []
        , br [] []
        , h4 [] [ span [ class "text-muted text-center" ] [ text "All folders" ] ]
        , MoveCopy.viewSearchBox SearchInput model.filter
        , MoveCopy.viewDirList DirChosen model.filter model.allDirs
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
            [ h4 [] [ text "Edit folders of »", text model.remote.name, text "«" ] ]
        |> Modal.body []
            [ Grid.containerFluid []
                [ Grid.row
                    [ Row.attrs [ style "min-width" "60vh", class "scrollable-modal-row" ]
                    ]
                    (viewRemoteFoldersContent model)
                ]
            ]
        |> Modal.footer []
            [ Button.button
                [ Button.outlinePrimary
                , Button.attrs [ onClick <| AnimateModal Modal.hiddenAnimated ]
                ]
                [ text "Close" ]
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
        ]
