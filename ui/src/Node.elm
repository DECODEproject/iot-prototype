module Node exposing (..)

import Http
import Html exposing (..)
import Html.Attributes exposing (..)
import Html.Events exposing (onClick)
import List.Extra exposing (..)
import Decoders


main : Program Never Model Msg
main =
    Html.program
        { init = init
        , view = view
        , update = update
        , subscriptions = subscriptions
        }


type alias Model =
    { accepted : Maybe Decoders.Entitlements
    , requested : Maybe Decoders.Entitlements
    , metadata : Maybe Decoders.Metadata
    }


initialModel : Model
initialModel =
    { accepted = Nothing
    , requested = Nothing
    , metadata = Nothing
    }


init : ( Model, Cmd Msg )
init =
    ( initialModel, getMetadata )


type Msg
    = NoOp
    | GetMetadataCompleted (Result Http.Error Decoders.Metadata)
    | GetAcceptedEntitlementsCompleted (Result Http.Error Decoders.Entitlements)
    | GetRequestedEntitlementsCompleted (Result Http.Error Decoders.Entitlements)


update : Msg -> Model -> ( Model, Cmd Msg )
update msg model =
    case msg of
        NoOp ->
            ( model, Cmd.none )

        GetAcceptedEntitlementsCompleted result ->
            case result of
                Err httpError ->
                    let
                        _ =
                            Debug.log "GetAcceptedEntitlementsCompleted error" httpError
                    in
                        ( model, Cmd.none )

                Ok items ->
                    ( { model | accepted = Just items }, getRequestedEntitlements )

        GetRequestedEntitlementsCompleted result ->
            case result of
                Err httpError ->
                    let
                        _ =
                            Debug.log "GetRequestedEntitlementsCompleted error" httpError
                    in
                        ( model, Cmd.none )

                Ok items ->
                    ( { model | requested = Just items }, Cmd.none )

        GetMetadataCompleted result ->
            case result of
                Err httpError ->
                    let
                        _ =
                            Debug.log "GetMetadataCompleted error" httpError
                    in
                        ( model, Cmd.none )

                Ok items ->
                    ( { model | metadata = Just items }, getAcceptedEntitlements )



--RPC


getMetadata : Cmd Msg
getMetadata =
    let
        url =
            "http://localhost:8080/data/meta"

        request =
            Http.get url Decoders.decodeMetadata
    in
        Http.send GetMetadataCompleted request


getAcceptedEntitlements : Cmd Msg
getAcceptedEntitlements =
    let
        url =
            "http://localhost:8080/entitlements/accepted/"

        request =
            Http.get url Decoders.decodeEntitlements
    in
        Http.send GetAcceptedEntitlementsCompleted request


getRequestedEntitlements : Cmd Msg
getRequestedEntitlements =
    let
        url =
            "http://localhost:8080/entitlements/requests/"

        request =
            Http.get url Decoders.decodeEntitlements
    in
        Http.send GetRequestedEntitlementsCompleted request



-- SUBSCRIPTIONS


subscriptions : Model -> Sub Msg
subscriptions model =
    Sub.none



-- VIEW


view : Model -> Html Msg
view model =
    case model.metadata of
        Nothing ->
            text ("no data exists.")

        Just e ->
            div []
                [ text ("Node")
                , div []
                    [ text ("Data")
                    , drawMetadata e model
                    , div [] [ text (toString model) ]
                    ]
                ]


drawEntitlements : Decoders.Entitlements -> Html Msg
drawEntitlements e =
    div [] <| List.map (\ent -> drawEntitlement ent) e


drawEntitlement : Decoders.Entitlement -> Html Msg
drawEntitlement e =
    div [] [ text (e.subject), text (":"), text (e.level) ]


drawMetadata : Decoders.Metadata -> Model -> Html Msg
drawMetadata e model =
    div [] <| List.map (\m -> drawMetadataItem m model) e


drawMetadataItem : Decoders.MetadataItem -> Model -> Html Msg
drawMetadataItem m model =
    div [] [ text (m.description), drawEntitlementSelector m model ]


drawEntitlementSelector : Decoders.MetadataItem -> Model -> Html Msg
drawEntitlementSelector m model =
    let
        accepted =
            findEntitlement m.key model.accepted

        requested =
            findEntitlement m.key model.requested
    in
        div [] [ text (" current : "), (drawAccepted accepted), drawRequested (requested) ]


drawAccepted : Maybe Decoders.Entitlement -> Html Msg
drawAccepted ent =
    case ent of
        Nothing ->
            text ("entitlement not set")

        Just e ->
            text (e.level)


drawRequested : Maybe Decoders.Entitlement -> Html Msg
drawRequested ent =
    case ent of
        Nothing ->
            text ("")

        Just e ->
            div [] [ text (" request : "), text (e.level), a [ href "#" ] [ text ("accept") ], a [ href "#" ] [ text ("decline") ] ]


findEntitlement : String -> Maybe Decoders.Entitlements -> Maybe Decoders.Entitlement
findEntitlement key entitlements =
    case entitlements of
        Nothing ->
            Nothing

        Just ents ->
            List.Extra.find (\e -> e.uid == key) ents
