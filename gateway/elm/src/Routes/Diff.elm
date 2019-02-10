module Routes.Diff exposing
    ( Model
    , Msg
    , newModel
    , reload
    , subscriptions
    , update
    , updateUrl
    , view
    )

import Bootstrap.Button as Button
import Bootstrap.Grid as Grid
import Bootstrap.Grid.Col as Col
import Bootstrap.Grid.Row as Row
import Browser.Navigation as Nav
import Commands
import Html exposing (..)
import Html.Attributes exposing (..)
import Html.Events exposing (..)
import Http
import Time
import Url
import Url.Parser exposing ((</>), parse, s, string)
import Util



-- MODEL:


type State
    = Loading
    | Finished (Result String Commands.Diff)


type alias Model =
    { key : Nav.Key
    , url : Url.Url
    , zone : Time.Zone
    , state : State
    }


newModel : Nav.Key -> Url.Url -> Time.Zone -> Model
newModel key url zone =
    { key = key
    , url = url
    , zone = zone
    , state = Loading
    }


updateUrl : Model -> Url.Url -> Model
updateUrl model url =
    { model | url = url }


nameFromUrl : Url.Url -> String
nameFromUrl url =
    Maybe.withDefault ""
        (parse (s "diff" </> string) url)


reload : Model -> Url.Url -> Cmd Msg
reload model url =
    Commands.doRemoteDiff GotResponse (nameFromUrl url)



-- MESSAGES:


type Msg
    = GotResponse (Result Http.Error Commands.Diff)
    | BackClicked



-- UPDATE:


update : Msg -> Model -> ( Model, Cmd Msg )
update msg model =
    case msg of
        GotResponse result ->
            case result of
                Ok diff ->
                    ( { model | state = Finished (Ok diff) }, Cmd.none )

                Err err ->
                    ( { model | state = Finished (Err (Util.httpErrorToString err)) }, Cmd.none )

        BackClicked ->
            ( model, Nav.back model.key 1 )



-- VIEW:


viewSingle : List Commands.Entry -> Html Msg -> Html Msg
viewSingle entries header =
    if List.length entries > 0 then
        span []
            [ header
            , span [] (List.map (\e -> text <| " " ++ e.path) entries)
            , br [] []
            , br [] []
            ]

    else
        text ""


viewPairs : List Commands.DiffPair -> Html Msg -> Html Msg
viewPairs entries header =
    if List.length entries > 0 then
        span []
            [ header
            , span [] (List.map (\p -> text (" " ++ p.src.path ++ " ↔ " ++ p.dst.path)) entries)
            , br [] []
            , br [] []
            ]

    else
        text ""


viewHeading : String -> String -> Html Msg
viewHeading className message =
    h5 [ class className ] [ text message ]


viewDiff : Model -> Commands.Diff -> Html Msg
viewDiff model diff =
    let
        nChanges =
            Commands.diffChangeCount diff
    in
    case nChanges of
        0 ->
            text "There are no differences!"

        n ->
            div []
                [ viewSingle diff.added (viewHeading "text-success" "Added")
                , viewSingle diff.removed (viewHeading "text-warning" "Removed")
                , viewSingle diff.ignored (viewHeading "text-muted" "Ignored")
                , viewSingle diff.missing (viewHeading "text-secondary" "Missing")
                , viewPairs diff.moved (viewHeading "text-primary" "Moved")
                , viewPairs diff.merged (viewHeading "text-info" "Merged")
                , viewPairs diff.conflict (viewHeading "text-danger" "Conflicts")
                , br [] []
                , br [] []
                , text (String.fromInt n ++ " changes in total")
                ]


viewDiffContainer : Model -> Result String Commands.Diff -> Html Msg
viewDiffContainer model result =
    Grid.row []
        [ Grid.col [ Col.lg2, Col.attrs [ class "d-none d-lg-block" ] ] []
        , Grid.col [ Col.lg8, Col.md12 ]
            [ h4 [ class "text-center" ]
                [ span [ class "text-muted" ] [ text "Difference to »" ]
                , text (nameFromUrl model.url)
                , span [ class "text-muted" ] [ text "«" ]
                , span [ class "text-muted" ] [ text "«" ]
                , Button.button
                    [ Button.roleLink
                    , Button.attrs [ onClick BackClicked ]
                    ]
                    [ span [ class "font-weight-light" ] [ text "(go back)" ] ]
                ]
            , br [] []
            , case result of
                Ok diff ->
                    viewDiff model diff

                Err err ->
                    text err
            , br [] []
            ]
        , Grid.col [ Col.lg2, Col.attrs [ class "d-none d-lg-block" ] ] []
        ]


view : Model -> Html Msg
view model =
    case model.state of
        Loading ->
            text "Still loading"

        Finished result ->
            Grid.row []
                [ Grid.col
                    [ Col.lg12 ]
                    [ Grid.row [ Row.attrs [ id "main-content-row" ] ]
                        [ Grid.col
                            [ Col.xl10 ]
                            [ viewDiffContainer model result ]
                        ]
                    ]
                ]



-- SUBSCRIPTIONS:


subscriptions : Model -> Sub Msg
subscriptions model =
    Sub.none
