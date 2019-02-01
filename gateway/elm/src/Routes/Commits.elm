module Routes.Commits exposing (Model, Msg, newModel, reload, subscriptions, update, view)

import Bootstrap.Button as Button
import Bootstrap.Form.Input as Input
import Bootstrap.Form.InputGroup as InputGroup
import Bootstrap.Grid as Grid
import Bootstrap.Grid.Col as Col
import Bootstrap.Grid.Row as Row
import Bootstrap.ListGroup as ListGroup
import Bootstrap.Table as Table
import Bootstrap.Text as Text
import Browser.Navigation as Nav
import Commands
import Html exposing (..)
import Html.Attributes exposing (..)
import Html.Events exposing (..)
import Html.Lazy as Lazy
import Http
import Time
import Util



-- MODEL:


type State
    = Loading
    | Failure String
    | Success (List Commands.Commit)


type alias Model =
    { key : Nav.Key
    , state : State
    , zone : Time.Zone
    , filter : String
    }


newModel : Nav.Key -> Time.Zone -> Model
newModel key zone =
    Model key Loading zone ""



-- MESSAGES:


type Msg
    = GotLogResponse (Result Http.Error (List Commands.Commit))
    | GotResetResponse (Result Http.Error String)
    | CheckoutClicked String
    | SearchInput String



-- UPDATE:


reload : Cmd Msg
reload =
    Commands.doLog GotLogResponse


update : Msg -> Model -> ( Model, Cmd Msg )
update msg model =
    case msg of
        GotLogResponse result ->
            case result of
                Ok commits ->
                    ( { model | state = Success commits }, Cmd.none )

                Err err ->
                    ( { model | state = Failure (Util.httpErrorToString err) }, Cmd.none )

        GotResetResponse result ->
            case result of
                Ok _ ->
                    -- TODO: Display message.
                    ( model, Cmd.none )

                Err err ->
                    -- TODO: Handle error.
                    ( model, Cmd.none )

        CheckoutClicked hash ->
            ( model, Commands.doReset GotResetResponse "/" hash )

        SearchInput filter ->
            ( { model | filter = filter }, Cmd.none )



-- VIEW:


viewSearchBox : Model -> Html Msg
viewSearchBox model =
    InputGroup.config
        (InputGroup.text
            [ Input.placeholder "Search"
            , Input.attrs
                [ onInput SearchInput
                , value model.filter
                ]
            ]
        )
        |> InputGroup.successors
            [ InputGroup.span [ class "input-group-addon" ]
                [ button [] [ span [ class "fas fa-search fa-xs input-group-addon" ] [] ]
                ]
            ]
        |> InputGroup.attrs [ class "stylish-input-group input-group" ]
        |> InputGroup.view


filterCommits : String -> List Commands.Commit -> List Commands.Commit
filterCommits filter commits =
    commits
        |> List.filter (\c -> String.length c.msg > 0)
        |> List.filter
            (\c ->
                if filter == "" then
                    True

                else
                    String.contains filter c.msg
            )


viewCommit : Model -> Commands.Commit -> ListGroup.Item Msg
viewCommit model commit =
    ListGroup.li []
        [ Grid.row
            []
            [ Grid.col
                [ Col.xs1
                , Col.textAlign Text.alignXsLeft
                ]
                [ span [ class "fas fa-lg fa-save text-xs-right" ] []
                ]
            , Grid.col [ Col.xs9, Col.textAlign Text.alignXsLeft ]
                [ text commit.msg
                ]
            , Grid.col
                [ Col.xs2
                , Col.textAlign Text.alignXsRight
                ]
                [ Button.button
                    [ Button.outlineDanger
                    , Button.attrs [ onClick <| CheckoutClicked commit.hash ]
                    ]
                    [ text "Checkout" ]
                ]
            ]
        ]


viewCommitList : Model -> List Commands.Commit -> Html Msg
viewCommitList model commits =
    ListGroup.ul (List.map (viewCommit model) (filterCommits model.filter commits))


viewCommitListContainer : Model -> List Commands.Commit -> Html Msg
viewCommitListContainer model commits =
    Grid.row []
        [ Grid.col [ Col.lg2, Col.attrs [ class "d-none d-lg-block" ] ] []
        , Grid.col [ Col.lg8, Col.md12 ]
            [ h4 [ class "text-muted text-center" ] [ text "Commits" ]
            , br [] []
            , viewCommitList model commits
            , br [] []
            ]
        , Grid.col [ Col.lg2, Col.attrs [ class "d-none d-lg-block" ] ] []
        ]


view : Model -> Html Msg
view model =
    case model.state of
        Loading ->
            text "Still loading"

        Failure err ->
            text ("Failed to load log: " ++ err)

        Success commits ->
            Grid.row []
                [ Grid.col
                    [ Col.lg12 ]
                    [ Grid.row [ Row.attrs [ id "main-header-row" ] ]
                        [ Grid.col [ Col.xl9 ] [ text "" ]
                        , Grid.col [ Col.xl3 ] [ Lazy.lazy viewSearchBox model ]
                        ]
                    , Grid.row [ Row.attrs [ id "main-content-row" ] ]
                        [ Grid.col
                            [ Col.xl10 ]
                            [ viewCommitListContainer model commits
                            ]
                        ]
                    ]
                ]



-- SUBSCRIPTIONS:


subscriptions : Model -> Sub Msg
subscriptions model =
    Sub.none
