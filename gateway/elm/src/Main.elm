module Main exposing (init, main, subscriptions, update, view)

import Bootstrap.Breadcrumb as Breadcrumb
import Bootstrap.Button as Button
import Bootstrap.Form.Checkbox as Checkbox
import Bootstrap.Form.Input as Input
import Bootstrap.Form.InputGroup as InputGroup
import Bootstrap.Grid as Grid
import Bootstrap.Grid.Col as Col
import Bootstrap.Grid.Row as Row
import Bootstrap.Modal as Modal
import Bootstrap.Navbar as Navbar
import Bootstrap.Progress as Progress
import Bootstrap.Table as Table
import Browser
import Browser.Navigation as Nav
import File
import Filesize
import Html exposing (..)
import Html.Attributes exposing (..)
import Html.Events exposing (..)
import Http
import Json.Decode as D
import Json.Encode as E
import List
import Login
import Ls
import Set
import Task
import Time
import Url
import Util exposing (..)



-- MAIN


main : Program () Model Msg
main =
    Browser.application
        { init = init
        , update = update
        , subscriptions = subscriptions
        , view = view
        , onUrlChange = UrlChanged
        , onUrlRequest = LinkClicked
        }



-- MODEL


type Msg
    = FetchList String
    | GotLsResp (Result Http.Error (List Ls.Entry))
    | GotLoginResp (Result Http.Error Bool)
    | AdjustTimeZone Time.Zone
    | LinkClicked Browser.UrlRequest
    | UrlChanged Url.Url
    | CheckboxTick String Bool
    | CheckboxTickAll Bool
    | ModalShowUpload
    | ModalCloseUpload
    | UploadSelectedFiles (List File.File)
    | UploadProgress Http.Progress
    | Uploaded (Result Http.Error ())
    | UploadCancel


type UploadState
    = UploadFileSelect
    | Uploading Float
    | UploadDone
    | UploadFail


type ListModel
    = LsFailure
    | LoginFailure
    | Loading
    | LoginSuccess
    | LsSuccess (List Ls.Entry)


type alias Model =
    { listState : ListModel
    , zone : Time.Zone
    , checkedEntries : Set.Set String
    , key : Nav.Key
    , url : Url.Url
    , uploadState : UploadState
    , uploadModalState : Modal.Visibility
    }


init : () -> Url.Url -> Nav.Key -> ( Model, Cmd Msg )
init _ url key =
    ( { listState = Loading
      , zone = Time.utc
      , key = key
      , checkedEntries = Set.empty
      , url = url
      , uploadState = UploadFileSelect
      , uploadModalState = Modal.hidden
      }
    , Cmd.batch
        [ doLoginQuery (Login.Query "ali" "ila")
        , Task.perform AdjustTimeZone Time.here
        ]
    )



-- UPDATE


splitPath : String -> List String
splitPath path =
    List.filter (\s -> String.length s > 0) (String.split "/" path)


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
        splitUrl =
            List.reverse (splitPath path)
    in
    case splitUrl of
        [] ->
            "/"

        x :: _ ->
            x


update : Msg -> Model -> ( Model, Cmd Msg )
update msg model =
    case msg of
        AdjustTimeZone newZone ->
            ( { model | zone = newZone }, Cmd.none )

        FetchList path ->
            ( { model | listState = Loading }, doLsQuery (Ls.Query path 1) )

        GotLsResp result ->
            let
                updatedModel =
                    case result of
                        Ok entries ->
                            { model | listState = LsSuccess entries }

                        Err _ ->
                            { model | listState = LsFailure }
            in
            -- When changing pages, we should clear the existing checkmarks.
            ( { updatedModel | checkedEntries = Set.empty }, Cmd.none )

        GotLoginResp result ->
            case result of
                Ok _ ->
                    -- Immediately hit off a list query, which will in turn populate
                    -- the list view. Take the path from the current URL.
                    ( { model | listState = LoginSuccess }
                    , doLsQuery (Ls.Query (urlToPath model.url) 1)
                    )

                Err _ ->
                    ( { model | listState = LoginFailure }, Cmd.none )

        LinkClicked urlRequest ->
            case urlRequest of
                Browser.Internal url ->
                    ( model, Nav.pushUrl model.key (Url.toString url) )

                Browser.External href ->
                    ( model, Nav.load href )

        UrlChanged url ->
            ( { model | url = url }
            , doLsQuery (Ls.Query (urlToPath url) 1)
            )

        CheckboxTick path isChecked ->
            case model.listState of
                LsSuccess entries ->
                    case isChecked of
                        True ->
                            let
                                updatedSet =
                                    Set.insert path model.checkedEntries
                            in
                            ( { model
                                | checkedEntries =
                                    if Set.size updatedSet == List.length entries then
                                        Set.insert "" updatedSet

                                    else
                                        updatedSet
                              }
                            , Cmd.none
                            )

                        False ->
                            ( { model
                                | checkedEntries =
                                    Set.remove "" <| Set.remove path model.checkedEntries
                              }
                            , Cmd.none
                            )

                _ ->
                    ( model, Cmd.none )

        CheckboxTickAll isChecked ->
            case isChecked of
                True ->
                    case model.listState of
                        LsSuccess entries ->
                            ( { model
                                | checkedEntries =
                                    Set.fromList
                                        (List.map (\e -> e.path) entries
                                            ++ [ "" ]
                                        )
                              }
                            , Cmd.none
                            )

                        _ ->
                            ( model, Cmd.none )

                False ->
                    ( { model | checkedEntries = Set.empty }, Cmd.none )

        ModalShowUpload ->
            ( { model | uploadModalState = Modal.shown, uploadState = UploadFileSelect }, Cmd.none )

        ModalCloseUpload ->
            ( { model | uploadModalState = Modal.hidden, uploadState = UploadFileSelect }, Cmd.none )

        UploadSelectedFiles files ->
            ( { model | uploadState = Uploading 0 }
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
                    ( { model | uploadState = Uploading <| Http.fractionSent p }, Cmd.none )

                Http.Receiving _ ->
                    ( model, Cmd.none )

        Uploaded result ->
            case result of
                Ok _ ->
                    ( { model | uploadState = UploadDone }, Cmd.none )

                Err _ ->
                    ( { model | uploadState = UploadFail }, Cmd.none )

        UploadCancel ->
            ( { model | uploadState = UploadFileSelect }, Cmd.none )



-- SUBSCRIPTIONS


subscriptions : Model -> Sub Msg
subscriptions model =
    Http.track "upload" UploadProgress



-- VIEW


view : Model -> Browser.Document Msg
view model =
    { title = "gateway"
    , body =
        [ Grid.containerFluid []
            [ Grid.row []
                [ Grid.col
                    [ Col.md2
                    , Col.attrs
                        [ class "d-none"
                        , class "d-md-block"
                        , class "bg-light"
                        , class "sidebar"
                        ]
                    ]
                    [ viewSidebar model ]
                , Grid.col
                    [ Col.lg10
                    , Col.attrs
                        [ class "ml-sm-auto"
                        , class "px-4"
                        , id "main-column"
                        ]
                    ]
                    [ Grid.row [ Row.attrs [ id "main-header-row" ] ]
                        [ Grid.col [ Col.xl9 ] [ viewBreadcrumbs model ]
                        , Grid.col [ Col.xl3 ] [ viewSearchBox model ]
                        ]
                    , Grid.row [ Row.attrs [ id "main-content-row" ] ]
                        [ Grid.col [ Col.xl10 ] [ viewLsResponse model ]
                        , Grid.col [ Col.xl2 ] [ viewActionList model ]
                        ]
                    ]
                ]
            , viewModalUpload model
            ]
        ]
    }


filesDecoder : D.Decoder (List File.File)
filesDecoder =
    D.at [ "target", "files" ] (D.list File.decoder)


showUploadState : Model -> List (Grid.Column Msg)
showUploadState model =
    case model.uploadState of
        UploadFileSelect ->
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
            -- TODO: Provide cancel button.
            [ Grid.col [ Col.xs10 ]
                [ Progress.progress
                    [ Progress.animated
                    , Progress.value (100 * fraction)
                    ]
                ]
            , Grid.col [ Col.xs2 ]
                [ Button.button [ Button.outlinePrimary, Button.attrs [ onClick UploadCancel ] ] [ text "Cancel" ]
                ]
            ]

        UploadDone ->
            [ Grid.col [ Col.xs12 ] [ text "Upload done" ] ]

        UploadFail ->
            [ Grid.col [ Col.xs12 ] [ text "Upload failed for unknown reasons" ] ]


viewModalUpload : Model -> Html Msg
viewModalUpload model =
    Modal.config ModalCloseUpload
        |> Modal.large
        |> Modal.h5 [] [ text "Upload a new file" ]
        |> Modal.body []
            [ Grid.containerFluid []
                [ Grid.row [] (showUploadState model) ]
            ]
        |> Modal.footer []
            [ Button.button
                [ Button.outlinePrimary
                , Button.attrs [ onClick ModalCloseUpload ]
                ]
                [ text "Close" ]
            ]
        |> Modal.view model.uploadModalState


buildBreadcrumbs : List String -> List String -> List (Breadcrumb.Item msg)
buildBreadcrumbs names previous =
    let
        displayName =
            \n ->
                if String.length n <= 0 then
                    "Home"

                else
                    n
    in
    case names of
        [] ->
            -- Recursion stop.
            []

        [ name ] ->
            -- Final element in the breadcrumbs.
            -- Already selected therefore.
            [ Breadcrumb.item []
                [ text (displayName name)
                ]
            ]

        name :: rest ->
            -- Some intermediate element.
            [ Breadcrumb.item []
                [ a [ href ("/view" ++ String.join "/" (previous ++ [ name ])) ]
                    [ text (displayName name) ]
                ]
            ]
                ++ buildBreadcrumbs rest (previous ++ [ name ])


viewBreadcrumbs : Model -> Html msg
viewBreadcrumbs model =
    div [ id "breadcrumbs-box" ]
        [ Breadcrumb.container
            (buildBreadcrumbs ([ "" ] ++ (urlToPath model.url |> splitPath)) [])
        ]


viewSearchBox : Model -> Html msg
viewSearchBox model =
    InputGroup.config
        (InputGroup.text
            [ Input.placeholder "Search" ]
        )
        |> InputGroup.successors
            [ InputGroup.span [ class "input-group-addon" ]
                [ button [] [ span [ class "fas fa-search fa-xs input-group-addon" ] [] ]
                ]
            ]
        |> InputGroup.attrs [ class "stylish-input-group input-group" ]
        |> InputGroup.view


viewSidebar : Model -> Html Msg
viewSidebar model =
    div [ class "sidebar-sticky" ]
        [ ul [ class "nav", class "flex-column" ]
            [ li [ class "nav-item" ]
                [ a [ class "nav-link active", href "/view" ]
                    [ span [ class "fas fa-4x fa-torii-gate" ] [] ]
                ]
            , br [] []
            , li [ class "nav-item" ]
                [ a [ class "nav-link", href "#" ]
                    [ span [ class "text-muted" ] [ text "Files" ] ]
                ]
            , li [ class "nav-item" ]
                [ a [ class "nav-link", href "#" ]
                    [ span [ class "text-muted" ] [ text "Remotes" ] ]
                ]
            , li [ class "nav-item" ]
                [ a [ class "nav-link", href "#" ]
                    [ span [ class "text-muted" ] [ text "Deleted files" ] ]
                ]
            , li [ class "nav-item" ]
                [ a [ class "nav-link", href "#" ]
                    [ span [ class "text-muted" ] [ text "Settings" ] ]
                ]
            ]
        , div [ id "sidebar-bottom" ]
            [ hr [] []
            , p [ id "sidebar-bottom-text", class "text-muted" ]
                [ text "Powered by brig Ⓒ 2015 ‒ 2019"
                , br [] []
                , a [ href "https://github.com/sahib/brig" ] [ text "Get the source code here" ]
                ]
            ]
        ]


viewEntryIcon : Ls.Entry -> Html Msg
viewEntryIcon entry =
    case entry.isDir of
        True ->
            span [ class "far fa-lg fa-folder text-xs-right file-list-icon" ] []

        False ->
            span [ class "far fa-lg fa-file text-xs-right file-list-icon" ] []


formatLastModified : Time.Zone -> Time.Posix -> String -> Html Msg
formatLastModified z t owner =
    let
        timestamp =
            String.join
                "/"
                [ Time.toDay z t |> String.fromInt
                , Time.toMonth z t |> monthToInt |> String.fromInt
                , Time.toYear z t |> String.fromInt
                ]
                ++ " "
                ++ String.join ":"
                    [ Time.toHour z t |> String.fromInt |> String.padLeft 2 '0'
                    , Time.toMinute z t |> String.fromInt |> String.padLeft 2 '0'
                    , Time.toSecond z t |> String.fromInt |> String.padLeft 2 '0'
                    ]
    in
    p [] [ text timestamp, span [ class "text-muted" ] [ text " by " ], text owner ]


makeCheckbox : Bool -> (Bool -> Msg) -> Html Msg
makeCheckbox isChecked msg =
    div [ class "checkbox" ]
        [ label []
            [ input [ type_ "checkbox", onCheck msg, checked isChecked ] []
            , span [ class "cr" ] [ i [ class "cr-icon fas fa-lg fa-check" ] [] ]
            ]
        ]


readCheckedState : Model -> String -> Bool
readCheckedState model path =
    Set.member path model.checkedEntries


entriesToHtml : Model -> List Ls.Entry -> Html Msg
entriesToHtml model entries =
    Table.table
        { options = [ Table.hover ]
        , thead =
            Table.simpleThead
                [ Table.th [] [ makeCheckbox (readCheckedState model "") CheckboxTickAll ]
                , Table.th [] [ span [ class "icon-column" ] [ text "" ] ]
                , Table.th [] [ span [ class "text-muted" ] [ text "Name" ] ]
                , Table.th [] [ span [ class "text-muted" ] [ text "Modified" ] ]
                , Table.th [] [ span [ class "text-muted" ] [ text "Size" ] ]
                ]
        , tbody =
            Table.tbody []
                (List.map
                    (\e ->
                        Table.tr
                            []
                            [ Table.td []
                                [ makeCheckbox (readCheckedState model e.path) (CheckboxTick e.path)
                                ]
                            , Table.td [ Table.cellAttr (class "icon-column") ] [ viewEntryIcon e ]
                            , Table.td [] [ a [ "/view" ++ e.path |> href ] [ text (basename e.path) ] ]
                            , Table.td [] [ formatLastModified model.zone e.lastModified e.user ]
                            , Table.td [] [ text (Filesize.format e.size) ]
                            ]
                    )
                    entries
                )
        }


viewLsResponse : Model -> Html Msg
viewLsResponse model =
    case model.listState of
        LoginFailure ->
            div []
                [ text "I could not login. Try different user / password" ]

        LsFailure ->
            div []
                [ button
                    [ onClick (FetchList "/"), style "display" "block" ]
                    [ text "Try again!" ]
                ]

        Loading ->
            text "Loading..."

        LoginSuccess ->
            div []
                [ button
                    [ onClick (FetchList "/"), style "display" "block" ]
                    [ text "Show list!" ]
                , text "you are logged in."
                ]

        LsSuccess entries ->
            div []
                [ entriesToHtml model entries ]


numberOfSelectedItems : Model -> Int
numberOfSelectedItems model =
    Set.filter (\e -> String.isEmpty e |> not) model.checkedEntries |> Set.size


labelSelectedItems : Int -> String
labelSelectedItems num =
    case num of
        0 ->
            "Nothing selected"

        1 ->
            "1 item selected"

        n ->
            String.fromInt n ++ " items selected"


viewActionList : Model -> Html Msg
viewActionList model =
    let
        nSelected =
            numberOfSelectedItems model
    in
    div [ class "toolbar" ]
        [ Button.button
            [ Button.primary
            , Button.block
            , Button.disabled (nSelected == 0)
            ]
            [ text "Share" ]
        , Button.button
            [ Button.primary
            , Button.block
            , Button.attrs [ onClick <| ModalShowUpload ]
            ]
            [ text "Upload" ]
        , br [] []
        , p [ class "text-muted" ] [ text (labelSelectedItems nSelected) ]
        , br [] []
        , ul [ id "toolbar-item-list" ]
            [ li [ class "nav-item" ]
                [ a [ class "nav-link", href "#" ]
                    [ span [ class "fas fa-lg fa-history" ] [], text " Version history" ]
                ]
            , li [ class "nav-item" ]
                [ a [ class "nav-link", href "#" ]
                    [ span [ class "fas fa-lg fa-edit" ] [], text " Create Folder" ]
                ]
            , li [ class "nav-item" ]
                [ a [ class "nav-link", href "#" ]
                    [ span [ class "fas fa-lg fa-arrow-right" ] [], text " Move" ]
                ]
            , li [ class "nav-item" ]
                [ a [ class "nav-link", href "#" ]
                    [ span [ class "fas fa-lg fa-copy" ] [], text " Copy" ]
                ]
            , li [ class "nav-item" ]
                [ a [ class "nav-link", href "#" ]
                    [ span [ class "fas fa-lg fa-trash" ] [], text " Delete" ]
                ]
            ]
        ]



-- QUERYING


doLoginQuery : Login.Query -> Cmd Msg
doLoginQuery q =
    Http.post
        { url = "/api/v0/login"
        , body = Http.jsonBody (Login.encode q)
        , expect = Http.expectJson GotLoginResp Login.decode
        }


doLsQuery : Ls.Query -> Cmd Msg
doLsQuery q =
    Http.post
        { url = "/api/v0/ls"
        , body = Http.jsonBody (Ls.encode q)
        , expect = Http.expectJson GotLsResp Ls.decode
        }
