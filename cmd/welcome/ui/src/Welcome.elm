port module Welcome exposing (..)

import Browser
import Html exposing (..)
import Html.Attributes exposing (..)
import Html.Events exposing (..)
import VitePluginHelper exposing (asset)



-- MAIN


main : Program () Model Msg
main =
    Browser.element
        { init = init
        , update = update
        , view = view
        , subscriptions = subscriptions
        }



-- PORTS


port sendGreetRequest : String -> Cmd msg


port greetReceiver : (String -> msg) -> Sub msg



-- MODEL


type alias Model =
    { screen : Int
    }


init : () -> ( Model, Cmd msg )
init _ =
    ( { screen = 0
      }
    , Cmd.none
    )



-- UPDATE


type Msg
    = Screen Int


update : Msg -> Model -> ( Model, Cmd Msg )
update msg model =
    case msg of
        Screen i ->
            ( { model | screen = i }, Cmd.none )



-- SUBSCRIPTIONS


subscriptions : Model -> Sub Msg
subscriptions _ =
    Sub.none



-- VIEW


logo : Html Msg
logo =
    div [ class "max-w-xs mx-auto h-56 w-56" ] [ img [ src <| asset "./assets/icon.png?inline" ] [] ]


step : { children : Html Msg, buttonText : String, onButtonClick : Msg } -> Html Msg
step { children, buttonText, onButtonClick } =
    div [ class "bg-base-200 min-h-screen w-full flex items-center justify-center" ]
        [ div [ class "p-6 flex min-h-screen flex-col items-center justify-between space-y-3" ]
            [ div [ class "flex-none text-center flex flex-col items-center" ] [ children ]
            , button
                [ class "btn btn-primary w-full flex-none", onClick onButtonClick ]
                [ text buttonText ]
            ]
        ]


view : Model -> Html Msg
view model =
    case model.screen of
        0 ->
            step
                { children =
                    div [ class "flex flex-col items-center space-y-3" ]
                        [ logo
                        , div [ class "text-center" ]
                            [ h1 [ class "text-3xl" ] [ text "Welcome to" ]
                            , h2 [ class "text-primary font-extrabold text-4xl" ] [ text "Pareto Security" ]
                            ]
                        , p [ class "text-sm text-justify text-content" ]
                            [ text "Pareto Security is an app that regularly checks your Mac's security configuration. It helps you take care of 20% of security tasks that prevent 80% of problems." ]
                        , label [ class "fieldset-label" ]
                            [ input [ type_ "checkbox", class "checkbox checkbox-xs checkbox-primary" ] []
                            , text " Automatically launch on system startup"
                            ]
                        ]
                , buttonText = "Get Started"
                , onButtonClick = Screen 1
                }

        1 ->
            step
                { children =
                    div [ class "flex flex-col items-center space-y-3" ]
                        [ logo
                        , h1 [ class "text-3xl" ] [ text "Done!" ]
                        , p [ class "text-sm text-justify text-content grow" ]
                            [ text "Pareto Security is now running in the background. You can find the app by looking for "
                            , img [ src <| asset "./assets/icon_black.svg?inline", class "h-6 w-6 inline-block" ] []
                            , text " in the tray."
                            ]
                        ]
                , buttonText = "Continue"
                , onButtonClick = Screen 1
                }

        _ ->
            div [] []
