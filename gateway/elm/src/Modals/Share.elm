module Modals.Share exposing (Model, Msg, newModel, show, subscriptions, update, view)

import Bootstrap.Button as Button
import Bootstrap.Grid as Grid
import Bootstrap.Grid.Col as Col
import Bootstrap.Grid.Row as Row
import Bootstrap.Modal as Modal
import Html exposing (..)
import Html.Attributes exposing (..)
import Html.Events exposing (..)
import Url


type alias Model =
    { paths : List String
    , modal : Modal.Visibility
    }


type Msg
    = ModalShow
    | AnimateModal Modal.Visibility
    | ModalClose



-- INIT


newModel : Model
newModel =
    { paths = []
    , modal = Modal.hidden
    }



-- UPDATE


update : Msg -> Model -> ( Model, Cmd Msg )
update msg model =
    case msg of
        AnimateModal visibility ->
            ( { model | modal = visibility }, Cmd.none )

        ModalShow ->
            ( { model | modal = Modal.shown }, Cmd.none )

        ModalClose ->
            ( { model | modal = Modal.hidden, paths = [] }, Cmd.none )



-- VIEW


viewShare : Url.Url -> Model -> List (Grid.Column Msg)
viewShare url model =
    [ Grid.col [ Col.xs12 ]
        [ p [] [ text "Use those links to share files with people that do not use brig." ]
        , p [] [ b [] [ text "Note:" ], text "They still need to authenticate themselves." ]
        , pre []
            [ text "lalala"
            ]
        ]
    ]


view : Model -> Url.Url -> Html Msg
view model url =
    Modal.config ModalClose
        |> Modal.large
        |> Modal.withAnimation AnimateModal
        |> Modal.h5 [] [ text "Share hyperlinks" ]
        |> Modal.body []
            [ Grid.containerFluid []
                [ Grid.row [] (viewShare url model) ]
            ]
        |> Modal.footer []
            [ Button.button
                [ Button.outlinePrimary
                , Button.attrs [ onClick <| AnimateModal Modal.hiddenAnimated ]
                ]
                [ text "Close" ]
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
        ]
