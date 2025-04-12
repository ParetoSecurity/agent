module Welcome exposing (..)

import Browser
import Html exposing (Html, div, h1, h2, h3, img, text)
import Html.Attributes exposing (alt, src, width)
import VitePluginHelper exposing (asset)



---- MODEL ----


type alias Model =
    {}


init : ( Model, Cmd Msg )
init =
    ( {}, Cmd.none )



---- UPDATE ----


type Msg
    = NoOp


update : Msg -> Model -> ( Model, Cmd Msg )
update _ model =
    ( model, Cmd.none )



---- VIEW ----


view : Model -> Html Msg
view _ =
    div []
        [ h1 [] [ text "Your Elm App is working!" ]
        , h3 [] [ text "With vite-plugin-helper" ]
        , img [ src <| asset "./assets/icon.png", width 128, alt "with vite-plugin-helper" ] []
        ]



---- PROGRAM ----


main : Program () Model Msg
main =
    Browser.application
        { view = \model -> { title = "Test", body = [ view model ] }
        , init = \_ _ _ -> ( {}, Cmd.none )
        , update = \_ model -> ( model, Cmd.none )
        , subscriptions = always Sub.none
        , onUrlRequest = \_ -> NoOp
        , onUrlChange = \_ -> NoOp
        }
