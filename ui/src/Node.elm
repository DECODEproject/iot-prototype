module Node exposing(..)

import Http
import Html exposing (..)
import Html.Attributes exposing (..)
import Html.Events exposing (onClick)

import Decoders

main : Program Never Model Msg
main =
    Html.program
        { init = init
        , view = view
        , update = update
        , subscriptions = subscriptions
        }


type alias Model = {
    accepted : Maybe Decoders.Entitlements
}

initialModel : Model
initialModel =
    {
        accepted = Nothing
    }


init : ( Model, Cmd Msg )
init =
    ( initialModel, getEntitlements )

type Msg
    = NoOp 
    | GetEntitlementsCompleted (Result Http.Error Decoders.Entitlements)


update : Msg -> Model -> ( Model, Cmd Msg )
update msg model =
    case msg of
        NoOp ->
            ( model, Cmd.none )
        GetEntitlementsCompleted result ->
              case result of
                Err httpError ->
                    let
                        _ =
                            Debug.log "GetEntitlementsCompleted error" httpError
                    in
                        ( model, Cmd.none )
                Ok items ->
                    ( {model | accepted = Just items }, Cmd.none )
--RPC

getEntitlements :  Cmd Msg
getEntitlements =
     let
        url =
            "http://localhost:8080/entitlements/accepted/"

        request =
            Http.get url Decoders.decodeEntitlements
    in
        Http.send GetEntitlementsCompleted request



-- SUBSCRIPTIONS


subscriptions : Model -> Sub Msg
subscriptions model =
    Sub.none



-- VIEW


view : Model -> Html Msg
view model =
    case model.accepted of
        Nothing -> text("no entitlements")
        Just e -> div []
                [ text("Node")
                , div[] [
                    text("Accepted")
                    , drawEntitlements e
                    , div [] [ text(toString model) ]
                    ]
                ]



drawEntitlements : Decoders.Entitlements -> Html Msg
drawEntitlements e =
    div[] <| List.map( \ent -> drawEntitlement ent ) e
 

drawEntitlement : Decoders.Entitlement -> Html Msg
drawEntitlement e =
    div[] [ text ( e.subject ), text(":"), text( e.level ) ]

