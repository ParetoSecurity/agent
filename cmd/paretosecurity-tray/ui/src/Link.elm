port module Link exposing (..)

import Browser
import Html exposing (..)
import Html.Attributes exposing (..)
import Html.Events exposing (..)
import VitePluginHelper exposing (asset)



-- MAIN


main : Program Flags Model Msg
main =
    Browser.element
        { init = init
        , update = update
        , view = view
        , subscriptions = subscriptions
        }



-- FLAGS


type alias Flags =
    { inviteId : String
    , host : Maybe String
    }



-- PORTS


port linkDevice : { inviteId : String, host : Maybe String } -> Cmd msg


port quitApp : () -> Cmd msg


port linkDeviceCallback : (String -> msg) -> Sub msg



-- MODEL


type Screen
    = LinkingScreen
    | SuccessScreen
    | ErrorScreen String


type alias Model =
    { screen : Screen
    , inviteId : String
    , host : Maybe String
    }


init : Flags -> ( Model, Cmd msg )
init flags =
    ( { screen = LinkingScreen
      , inviteId = flags.inviteId
      , host = flags.host
      }
    , linkDevice { inviteId = flags.inviteId, host = flags.host }
    )



-- UPDATE


type Msg
    = LinkCallback String
    | Quit


update : Msg -> Model -> ( Model, Cmd Msg )
update msg model =
    case msg of
        LinkCallback result ->
            if result == "ok" then
                ( { model | screen = SuccessScreen }
                , Cmd.none
                )

            else
                ( { model | screen = ErrorScreen result }
                , Cmd.none
                )

        Quit ->
            ( model
            , quitApp ()
            )



-- SUBSCRIPTIONS


subscriptions : Model -> Sub Msg
subscriptions _ =
    linkDeviceCallback LinkCallback



-- VIEW


view : Model -> Html Msg
view model =
    case model.screen of
        LinkingScreen ->
            viewLinking model

        SuccessScreen ->
            viewSuccess

        ErrorScreen error ->
            viewError error


viewLinking : Model -> Html Msg
viewLinking model =
    div [ class "bg-base-200 min-h-screen w-full flex items-center justify-center" ]
        [ div [ class "p-4 pt-8 flex min-h-screen flex-col items-center justify-between space-y-3" ]
            [ div [ class "flex-1 flex items-center justify-center" ]
                [ div [ class "flex flex-col items-center space-y-8" ]
                    [ -- Logo
                      div [ class "max-w-xs mx-auto h-56 w-56" ]
                        [ img
                            [ src (asset "/src/assets/icon.png")
                            , alt "Pareto Security"
                            ]
                            []
                        ]

                    -- Add extra spacing after logo
                    , div [ class "pb-4" ] []

                    -- Progress section with candy spinner
                    , div [ class "w-full flex flex-col items-center space-y-3" ]
                        [ span [ class "loading loading-ring loading-lg" ] []
                        , p [ class "text-sm text-center text-content mt-2" ]
                            [ text "Linking your device to the team..." ]
                        ]

                    -- Invite ID display
                    , div [ class "text-sm text-content/60 text-center" ]
                        [ p [ class "font-mono break-all" ]
                            [ text ("ID: " ++ model.inviteId) ]
                        ]
                    ]
                ]
            , text "" -- No button during linking
            ]
        ]


viewSuccess : Html Msg
viewSuccess =
    div [ class "bg-base-200 min-h-screen w-full flex items-center justify-center" ]
        [ div [ class "p-4 pt-8 flex min-h-screen flex-col items-center justify-between space-y-3" ]
            [ div [ class "flex-1 flex items-center justify-center" ]
                [ div [ class "flex flex-col items-center space-y-3" ]
                    [ -- Logo
                      div [ class "max-w-xs mx-auto h-56 w-56" ]
                        [ img
                            [ src (asset "/src/assets/icon.png")
                            , alt "Pareto Security"
                            ]
                            []
                        ]

                    -- Add extra spacing after logo
                    , div [ class "pb-6" ] []

                    , h1 [ class "text-3xl" ] [ text "Success!" ]
                    , p [ class "text-sm text-justify text-content grow" ]
                        [ text "Your device has been successfully linked to the team. Pareto Security is now running in the background and will monitor your security settings." ]
                    ]
                ]
            , button
                [ class "btn btn-primary w-full flex-none"
                , onClick Quit
                ]
                [ text "Finish" ]
            ]
        ]


viewError : String -> Html Msg
viewError error =
    div [ class "bg-base-200 min-h-screen w-full flex items-center justify-center" ]
        [ div [ class "p-4 pt-8 flex min-h-screen flex-col items-center justify-between space-y-3" ]
            [ div [ class "flex-1 flex items-center justify-center" ]
                [ div [ class "flex flex-col items-center space-y-3" ]
                    [ -- Logo
                      div [ class "max-w-xs mx-auto h-56 w-56" ]
                        [ img
                            [ src (asset "/src/assets/icon.png")
                            , alt "Pareto Security"
                            ]
                            []
                        ]

                    -- Add extra spacing after logo
                    , div [ class "pb-6" ] []

                    , h1 [ class "text-3xl text-error" ] [ text "Error!" ]
                    , div [ class "text-center" ]
                        [ p [ class "text-sm text-content/70" ] [ text error ]
                        ]
                    , p [ class "text-sm text-justify text-content mt-3" ]
                        [ text "Please check your invitation link and try again." ]
                    ]
                ]
            , button
                [ class "btn btn-primary w-full flex-none"
                , onClick Quit
                ]
                [ text "Close" ]
            ]
        ]
