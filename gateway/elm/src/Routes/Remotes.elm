module Routes.Remotes exposing
    ( Model
    , Msg
    , buildModals
    , newModel
    , reload
    , subscriptions
    , update
    , view
    )

import Bootstrap.Alert as Alert
import Bootstrap.Button as Button
import Bootstrap.Dropdown as Dropdown


import Bootstrap.Grid as Grid
import Bootstrap.Grid.Col as Col
import Bootstrap.Grid.Row as Row
import Bootstrap.ListGroup as ListGroup
import Bootstrap.Table as Table
import Bootstrap.Text as Text
import Browser.Navigation as Nav
import Commands
import Delay
import Dict
import Html exposing (..)
import Html.Attributes exposing (..)
import Html.Events exposing (..)

import Http
import Modals.RemoteAdd as RemoteAdd
import Modals.RemoteFolders as RemoteFolders
import Modals.RemoteRemove as RemoteRemove
import Time
import Url
import Util



-- MODEL:


type State
    = Loading
    | Failure String
    | Success (List Commands.Remote)


type alias AlertState =
    { message : String
    , typ : Alert.Config Msg -> Alert.Config Msg
    , vis : Alert.Visibility
    }


defaultAlertState : AlertState
defaultAlertState =
    { message = ""
    , typ = Alert.danger
    , vis = Alert.closed
    }


type alias Model =
    { key : Nav.Key
    , zone : Time.Zone
    , state : State
    , self : Commands.Self
    , alert : AlertState
    , remoteAddState : RemoteAdd.Model
    , remoteRemoveState : RemoteRemove.Model
    , remoteFoldersState : RemoteFolders.Model
    , dropdowns : Dict.Dict String Dropdown.State
    }


newModel : Nav.Key -> Time.Zone -> Model
newModel key zone =
    { key = key
    , zone = zone
    , state = Loading
    , self = Commands.emptySelf
    , remoteAddState = RemoteAdd.newModel
    , remoteRemoveState = RemoteRemove.newModel
    , remoteFoldersState = RemoteFolders.newModel
    , dropdowns = Dict.empty
    , alert = defaultAlertState
    }



-- MESSAGES:


type Msg
    = GotRemoteListResponse (Result Http.Error (List Commands.Remote))
    | GotSyncResponse (Result Http.Error String)
    | GotSelfResponse (Result Http.Error Commands.Self)
    | GotRemoteModifyResponse (Result Http.Error String)
    | SyncClicked String
    | AutoUpdateToggled Commands.Remote Bool
      -- Sub messages:
    | RemoteAddMsg RemoteAdd.Msg
    | RemoteRemoveMsg RemoteRemove.Msg
    | RemoteFolderMsg RemoteFolders.Msg
    | DropdownMsg String Dropdown.State
    | AlertMsg Alert.Visibility



-- UPDATE:


reload : Cmd Msg
reload =
    Cmd.batch
        [ Commands.doRemoteList GotRemoteListResponse
        , Commands.doSelfQuery GotSelfResponse
        ]


showAlert : Model -> Float -> (Alert.Config Msg -> Alert.Config Msg) -> String -> ( Model, Cmd Msg )
showAlert model duration modalTyp message =
    let
        newAlert =
            AlertState message modalTyp Alert.shown
    in
    ( { model | alert = newAlert }
    , Cmd.batch
        [ Delay.after duration Delay.Second (AlertMsg Alert.closed) ]
    )


update : Msg -> Model -> ( Model, Cmd Msg )
update msg model =
    case msg of
        GotRemoteListResponse result ->
            case result of
                Ok remotes ->
                    ( { model | state = Success remotes }, Cmd.none )

                Err err ->
                    ( { model | state = Failure (Util.httpErrorToString err) }, Cmd.none )

        GotSyncResponse result ->
            case result of
                Ok _ ->
                    showAlert model 5 Alert.success "Succesfully synchronized!"

                Err err ->
                    showAlert model 20 Alert.danger ("Failed to sync: " ++ Util.httpErrorToString err)

        GotRemoteModifyResponse result ->
            case result of
                Ok _ ->
                    ( model, Cmd.none )

                Err err ->
                    showAlert model 20 Alert.danger ("Failed to set auto update: " ++ Util.httpErrorToString err)

        GotSelfResponse result ->
            case result of
                Ok self ->
                    ( { model | self = self }, Cmd.none )

                Err err ->
                    -- TODO: show failure message.
                    ( model, Cmd.none )

        DropdownMsg name state ->
            ( { model | dropdowns = Dict.insert name state model.dropdowns }, Cmd.none )

        SyncClicked name ->
            ( model, Commands.doRemoteSync GotSyncResponse name )

        AutoUpdateToggled remote state ->
            ( model, Commands.doRemoteModify GotRemoteModifyResponse { remote | acceptAutoUpdates = state } )

        RemoteAddMsg subMsg ->
            let
                ( upModel, upCmd ) =
                    RemoteAdd.update subMsg model.remoteAddState
            in
            ( { model | remoteAddState = upModel }, Cmd.map RemoteAddMsg upCmd )

        RemoteRemoveMsg subMsg ->
            let
                ( upModel, upCmd ) =
                    RemoteRemove.update subMsg model.remoteRemoveState
            in
            ( { model | remoteRemoveState = upModel }, Cmd.map RemoteRemoveMsg upCmd )

        RemoteFolderMsg subMsg ->
            let
                ( upModel, upCmd ) =
                    RemoteFolders.update subMsg model.remoteFoldersState
            in
            ( { model | remoteFoldersState = upModel }, Cmd.map RemoteFolderMsg upCmd )

        AlertMsg vis ->
            let
                newAlert =
                    AlertState model.alert.message model.alert.typ vis
            in
            ( { model | alert = newAlert }, Cmd.none )



-- VIEW:


viewAlert : AlertState -> Bool -> Html Msg
viewAlert alert isSuccess =
    Alert.config
        |> Alert.dismissableWithAnimation AlertMsg
        |> alert.typ
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
                    , text (" " ++ alert.message)
                    ]
                , Grid.col [ Col.xs2, Col.textAlign Text.alignXsRight ]
                    [ Button.button
                        [ Button.roleLink
                        , Button.attrs
                            [ class "notification-close-btn"
                            , onClick (AlertMsg Alert.closed)
                            ]
                        ]
                        [ span [ class "fas fa-xs fa-times" ] [] ]
                    ]
                ]
            ]
        |> Alert.view alert.vis


viewAutoUpdatesIcon : Bool -> Commands.Remote -> Html Msg
viewAutoUpdatesIcon state remote =
    Util.viewToggleSwitch (AutoUpdateToggled remote) "" state


viewRemoteState : Model -> Commands.Remote -> Html Msg
viewRemoteState model remote =
    if remote.isAuthenticated then
        if remote.isOnline then
            span [ class "fas fa-md fa-circle text-success" ] []

        else
            span [ class "text-warning" ]
                [ text <| Util.formatLastModified model.zone remote.lastSeen ]

    else
        span [ class "text-danger" ] [ text "not authenticated" ]


viewFingerprint : String -> Html Msg
viewFingerprint fingerprint =
    String.split ":" fingerprint
        |> List.map (\t -> String.slice 0 10 t)
        |> String.join ":"
        |> text


viewDropdown : Model -> Commands.Remote -> Html Msg
viewDropdown model remote =
    Dropdown.dropdown
        (Maybe.withDefault Dropdown.initialState (Dict.get remote.name model.dropdowns))
        { options = [ Dropdown.alignMenuRight ]
        , toggleMsg = DropdownMsg remote.name
        , toggleButton =
            Dropdown.toggle
                [ Button.roleLink ]
                [ span [ class "fas fa-ellipsis-h" ] [] ]
        , items =
            [ Dropdown.buttonItem
                [ onClick (SyncClicked remote.name)
                , disabled (not remote.isAuthenticated)
                ]
                [ span [ class "fas fa-md fa-sync-alt" ] [], text " Sync" ]
            , Dropdown.anchorItem
                [ disabled (not remote.isAuthenticated)
                , if remote.isAuthenticated then
                    href ("/diff/" ++ Url.percentEncode remote.name)

                  else
                    class "text-muted"
                ]
                [ span [ class "fas fa-md fa-search-minus" ] [], text " Diff" ]
            , Dropdown.divider
            , Dropdown.buttonItem
                [ onClick (RemoteRemoveMsg <| RemoteRemove.show remote.name) ]
                [ span [ class "text-danger" ]
                    [ span [ class "fas fa-md fa-times" ] []
                    , text " Remove"
                    ]
                ]
            ]
        }


viewRemote : Model -> Commands.Remote -> Table.Row Msg
viewRemote model remote =
    Table.tr []
        [ Table.td
            []
            [ span [ class "fas fa-lg fa-user-circle text-xs-right" ] [] ]
        , Table.td
            []
            [ text <| " " ++ remote.name ]
        , Table.td
            []
            [ viewRemoteState model remote ]
        , Table.td
            []
            [ span [ class "text-muted" ] [ viewFingerprint remote.fingerprint ] ]
        , Table.td
            []
            [ viewAutoUpdatesIcon remote.acceptAutoUpdates remote ]
        , Table.td
            []
            [ Button.button
                [ Button.roleLink
                , Button.attrs
                    [ onClick <| RemoteFolderMsg (RemoteFolders.show remote)
                    ]
                ]
                [ span
                    []
                    [ case List.length remote.folders of
                        0 ->
                            span [ class "fas fa-xs fa-asterisk" ] []

                        n ->
                            text <| String.fromInt n
                    ]
                ]
            ]
        , Table.td
            [ Table.cellAttr (class "text-right") ]
            [ viewDropdown model remote ]
        ]


viewRemoteList : Model -> List Commands.Remote -> Html Msg
viewRemoteList model remotes =
    Table.table
        { options =
            [ Table.hover
            , Table.attr (class "borderless-table")
            ]
        , thead =
            Table.thead []
                [ Table.tr []
                    [ Table.th
                        [ Table.cellAttr (style "width" "5%") ]
                        [ text "" ]
                    , Table.th
                        [ Table.cellAttr (style "width" "20%") ]
                        [ span [ class "text-muted" ] [ text "Name" ] ]
                    , Table.th
                        [ Table.cellAttr (style "width" "20%") ]
                        [ span [ class "text-muted" ] [ text "Online" ] ]
                    , Table.th
                        [ Table.cellAttr (style "width" "30%") ]
                        [ span [ class "text-muted" ] [ text "Fingerprint" ] ]
                    , Table.th
                        [ Table.cellAttr (style "width" "10%") ]
                        [ span [ class "text-muted" ] [ text "Auto Update" ] ]
                    , Table.th
                        [ Table.cellAttr (style "width" "10%") ]
                        [ span [ class "text-muted" ] [ text "Folders" ] ]
                    , Table.th
                        [ Table.cellAttr (style "width" "5%") ]
                        []
                    ]
                ]
        , tbody =
            Table.tbody []
                (List.map
                    (viewRemote model)
                    remotes
                )
        }


viewMetaRow : String -> Html msg -> Html msg
viewMetaRow key value =
    Grid.row []
        [ Grid.col [ Col.xs4, Col.textAlign Text.alignXsLeft ] [ span [ class "text-muted" ] [ text key ] ]
        , Grid.col [ Col.xs8, Col.textAlign Text.alignXsRight ] [ value ]
        ]


viewSelf : Model -> Html Msg
viewSelf model =
    Grid.row []
        [ Grid.col [ Col.lg2, Col.attrs [ class "d-none d-lg-block" ] ] []
        , Grid.col [ Col.xs12, Col.lg8, Col.textAlign Text.alignXsCenter ]
            [ ListGroup.ul
                [ ListGroup.li []
                    [ viewMetaRow "Name" (text model.self.name)
                    ]
                , ListGroup.li []
                    [ viewMetaRow "Fingerprint" (viewFingerprint model.self.fingerprint)
                    ]
                ]
            ]
        , Grid.col [ Col.lg2, Col.attrs [ class "d-none d-lg-block" ] ] []
        ]


viewRemoteListContainer : Model -> List Commands.Remote -> Html Msg
viewRemoteListContainer model remotes =
    Grid.row []
        [ Grid.col [ Col.lg1, Col.attrs [ class "d-none d-lg-block" ] ] []
        , Grid.col [ Col.xs12, Col.lg10 ]
            [ viewAlert model.alert True
            , viewRemoteList model remotes
            , div [ class "text-left" ]
                [ Button.button
                    [ Button.roleLink
                    , Button.attrs
                        [ onClick <| RemoteAddMsg RemoteAdd.show ]
                    ]
                    [ span [ class "fas fa-lg fa-plus" ] []
                    , text " Add new"
                    ]
                ]
            ]
        , Grid.col [ Col.lg1, Col.attrs [ class "d-none d-lg-block" ] ] []
        ]


view : Model -> Html Msg
view model =
    case model.state of
        Loading ->
            text "Still loading"

        Failure err ->
            text ("Failed to load log: " ++ err)

        Success remotes ->
            Grid.row []
                [ Grid.col
                    [ Col.lg12 ]
                    [ Grid.row [ Row.attrs [ id "main-header-row" ] ]
                        []
                    , Grid.row [ Row.attrs [ id "main-content-row" ] ]
                        [ Grid.col
                            [ Col.xl10 ]
                            [ h4 [ class "text-center text-muted" ] [ text "Own data" ]
                            , br [] []
                            , viewSelf model
                            , br [] []
                            , br [] []
                            , br [] []
                            , br [] []
                            , h4 [ class "text-center text-muted" ] [ text "Other remotes" ]
                            , br [] []
                            , viewRemoteListContainer model remotes
                            ]
                        ]
                    ]
                ]


buildModals : Model -> Html Msg
buildModals model =
    span []
        [ Html.map RemoteAddMsg (RemoteAdd.view model.remoteAddState)
        , Html.map RemoteRemoveMsg (RemoteRemove.view model.remoteRemoveState)
        , Html.map RemoteFolderMsg (RemoteFolders.view model.remoteFoldersState)
        ]



-- SUBSCRIPTIONS:


subscriptions : Model -> Sub Msg
subscriptions model =
    Sub.batch
        [ Alert.subscriptions model.alert.vis AlertMsg
        , Sub.map RemoteAddMsg <| RemoteAdd.subscriptions model.remoteAddState
        , Sub.map RemoteRemoveMsg <| RemoteRemove.subscriptions model.remoteRemoveState
        , Sub.map RemoteFolderMsg <| RemoteFolders.subscriptions model.remoteFoldersState
        , Sub.batch
            (List.map
                (\( name, state ) -> Dropdown.subscriptions state (DropdownMsg name))
                (Dict.toList model.dropdowns)
            )
        ]
