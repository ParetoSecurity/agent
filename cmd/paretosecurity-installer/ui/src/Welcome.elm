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


port installApp : Bool -> Cmd msg


port quitApp : () -> Cmd msg


port installAppCallback : (String -> msg) -> Sub msg



-- MODEL


type Screen
    = WelcomeScreen
    | InstallingScreen
    | ErrorScreen
    | DoneScreen


type alias Model =
    { screen : Screen
    , withStartup : Bool
    , message : Maybe String
    }


init : () -> ( Model, Cmd msg )
init _ =
    ( { screen = WelcomeScreen, withStartup = True, message = Nothing }
    , Cmd.none
    )



-- UPDATE


type Msg
    = Screen Screen
    | InstallCallback String
    | WithStartup Bool
    | Quit


update : Msg -> Model -> ( Model, Cmd Msg )
update msg model =
    case msg of
        Screen screen ->
            ( { model | screen = screen }
            , case screen of
                InstallingScreen ->
                    installApp model.withStartup

                _ ->
                    Cmd.none
            )

        InstallCallback m ->
            ( { model
                | screen = DoneScreen
                , message =
                    if m /= "ok" then
                        Just m

                    else
                        Nothing
              }
            , Cmd.none
            )

        WithStartup b ->
            ( { model | withStartup = b }
            , Cmd.none
            )

        Quit ->
            ( model, quitApp () )



-- SUBSCRIPTIONS


subscriptions : Model -> Sub Msg
subscriptions _ =
    installAppCallback InstallCallback



-- VIEW


logo : Html Msg
logo =
    div [ class "max-w-xs mx-auto h-56 w-56" ] [ img [ src <| asset "./assets/icon.png?inline" ] [] ]


step : { children : Html Msg, buttonText : String, onButtonClick : Msg } -> Html Msg
step { children, buttonText, onButtonClick } =
    div [ class "bg-base-200 min-h-screen w-full flex items-center justify-center" ]
        [ div [ class "p-4 pt-8 flex min-h-screen flex-col items-center justify-between space-y-3" ]
            [ div [ class "flex-1 flex items-center justify-center" ] [ children ]
            , if buttonText /= "" then
                button
                    [ class "btn btn-primary w-full flex-none", onClick onButtonClick ]
                    [ text buttonText ]

              else
                text ""
            ]
        ]


view : Model -> Html Msg
view model =
    case model.screen of
        WelcomeScreen ->
            step
                { children =
                    div [ class "flex flex-col items-center space-y-3" ]
                        [ logo
                        , div [ class "text-center" ]
                            [ h1 [ class "text-3xl" ] [ text "Welcome to" ]
                            , h2 [ class "text-primary font-extrabold text-4xl" ] [ text "Pareto Security" ]
                            ]
                        , p [ class "text-sm text-justify text-content" ]
                            [ text "Pareto Security is an app that regularly checks your security configuration. It helps you take care of 20% of security tasks that prevent 80% of problems." ]
                        , label [ class "fieldset-label text-sm" ]
                            [ input
                                [ type_ "checkbox"
                                , class "checkbox checkbox-xs checkbox-primary"
                                , checked model.withStartup
                                , onCheck WithStartup
                                ]
                                []
                            , text "Launch on system startup"
                            ]
                        ]
                , buttonText = "Get Started"
                , onButtonClick = Screen InstallingScreen
                }

        InstallingScreen ->
            step
                { children =
                    div [ class "flex flex-col items-center space-y-6" ]
                        [ logo
                        , div [ class "w-full flex flex-col items-center space-y-3" ]
                            [ node "progress"
                                [ class "progress w-full" ]
                                []
                            , p [ class "text-sm text-center text-content mt-2" ]
                                [ text "Installing Pareto Security app..." ]
                            ]
                        ]
                , buttonText = ""
                , onButtonClick = Screen DoneScreen
                }

        DoneScreen ->
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
                , buttonText = "Finish"
                , onButtonClick = Quit
                }

        ErrorScreen ->
            step
                { children =
                    div [ class "flex flex-col items-center space-y-3" ]
                        [ logo
                        , h1 [ class "text-3xl" ] [ text "Error!" ]
                        , p [ class "text-sm text-justify text-content grow" ]
                            [ text "An error occurred while installing Pareto Security. Please try again."
                            ]
                        , textarea
                            [ class "textarea textarea-bordered w-full max-w-xs"
                            , placeholder "Error message"
                            , readonly True
                            ]
                            [ text <| Maybe.withDefault "" model.message ]
                        ]
                , buttonText = "Retry"
                , onButtonClick = Screen InstallingScreen
                }
