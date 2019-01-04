module Main exposing (init, main, subscriptions, update, view)

import Bootstrap.Breadcrumb as Breadcrumb
import Bootstrap.Button as Button
import Bootstrap.Form.Checkbox as Checkbox
import Bootstrap.Form.Input as Input
import Bootstrap.Form.InputGroup as InputGroup
import Bootstrap.Grid as Grid
import Bootstrap.Grid.Col as Col
import Bootstrap.Grid.Row as Row
import Bootstrap.Navbar as Navbar
import Bootstrap.Table as Table
import Browser
import Browser.Navigation as Nav
import Html exposing (..)
import Html.Attributes exposing (..)
import Html.Events exposing (..)
import Http
import Json.Decode as D
import Json.Encode as E
import List
import Login
import Ls
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


type ListModel
    = LsFailure
    | LoginFailure
    | Loading
    | LoginSuccess
    | LsSuccess (List Ls.Entry)


type alias Model =
    { listState : ListModel
    , zone : Time.Zone
    , key : Nav.Key
    , url : Url.Url
    }


init : () -> Url.Url -> Nav.Key -> ( Model, Cmd Msg )
init _ url key =
    ( { listState = Loading
      , zone = Time.utc
      , key = key
      , url = url
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
    case splitPath url.path of
        [] ->
            "/"

        _ :: xs ->
            "/" ++ String.join "/" xs


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
            case result of
                Ok entries ->
                    ( { model | listState = LsSuccess entries }, Cmd.none )

                Err _ ->
                    ( { model | listState = LsFailure }, Cmd.none )

        GotLoginResp result ->
            case result of
                Ok _ ->
                    -- Immediately hit off a list query, which will in turn populate
                    -- the list view.
                    ( { model | listState = LoginSuccess }, doLsQuery (Ls.Query "/" 1) )

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



-- SUBSCRIPTIONS


subscriptions : Model -> Sub Msg
subscriptions model =
    Sub.none



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
            ]
        ]
    }


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
                [ a [ class "nav-link active", href "#" ]
                    [ span [ class "fas fa-3x fa-balance-scale" ] [] ]
                ]
            , br [] []
            , li [ class "nav-item" ]
                [ a [ class "nav-link", href "#" ]
                    [ span [ class "text-muted" ] [ text "Files" ] ]
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


formatLastModifiedTime : Time.Zone -> Time.Posix -> String
formatLastModifiedTime z t =
    String.join "/"
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


entriesToHtml : Model -> List Ls.Entry -> Html Msg
entriesToHtml model entries =
    Table.table
        { options = [ Table.hover ]
        , thead =
            Table.simpleThead
                [ Table.th [] []
                , Table.th [] [ span [ class "icon-column" ] [ text "" ] ]
                , Table.th [] [ span [ class "text-muted" ] [ text "Name" ] ]
                , Table.th [] [ span [ class "text-muted" ] [ text "Modified" ] ]
                , Table.th [] [ span [ class "text-muted" ] [ text "Owner" ] ]
                ]
        , tbody =
            Table.tbody []
                (List.map
                    (\e ->
                        Table.tr
                            []
                            [ Table.td []
                                [ div [ class "checkbox" ]
                                    [ label []
                                        [ input [ type_ "checkbox", value "" ] []
                                        , span [ class "cr" ] [ i [ class "cr-icon fas fa-lg fa-check" ] [] ]
                                        ]
                                    ]
                                ]
                            , Table.td [ Table.cellAttr (class "icon-column") ] [ viewEntryIcon e ]
                            , Table.td [] [ a [ "/view" ++ e.path |> href ] [ text (basename e.path) ] ]
                            , Table.td [] [ text (formatLastModifiedTime model.zone e.lastModified) ]
                            , Table.td [] [ text e.user ]
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


viewActionList : Model -> Html Msg
viewActionList model =
    div [ class "toolbar" ]
        [ p [ class "text-muted" ] [ text "1 item selected" ]
        , Button.button [ Button.primary, Button.block ] [ text "Share" ]
        , Button.button [ Button.primary, Button.block ] [ text "Upload" ]
        , br [] []
        , br [] []
        , ul [ id "toolbar-item-list" ]
            [ li [ class "nav-item" ]
                [ a [ class "nav-link", href "#" ]
                    [ span [ class "fas fa-lg fa-eye" ] [], text " View" ]
                ]
            , li [ class "nav-item" ]
                [ a [ class "nav-link", href "#" ]
                    [ span [ class "fas fa-lg fa-history" ] [], text " Version history" ]
                ]
            , li [ class "nav-item" ]
                [ a [ class "nav-link", href "#" ]
                    [ span [ class "fas fa-lg fa-edit" ] [], text " Rename" ]
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
