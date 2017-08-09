module Node exposing (..)

import Http
import Html exposing (..)
import Html.Attributes exposing (..)
import Html.Events exposing (onClick)
import List.Extra exposing (..)
import Decoders
import Json.Encode exposing (..)


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
    | AcceptEntitlement Decoders.Entitlement
    | AcceptEntitlementCompleted (Result Http.Error Decoders.Entitlement)
    | DeclineEntitlement Decoders.Entitlement
    | DeclineEntitlementCompleted (Result Http.Error Decoders.Entitlement)
    | AmendEntitlement Decoders.Entitlement
    | AmendEntitlementCompleted (Result Http.Error Decoders.Entitlement)


update : Msg -> Model -> ( Model, Cmd Msg )
update msg model =
    case msg of
        NoOp ->
            ( model, Cmd.none )

        GetAcceptedEntitlementsCompleted (Ok items) ->
            ( { model | accepted = Just items }, getRequestedEntitlements )

        GetAcceptedEntitlementsCompleted (Err httpError) ->
            Debug.crash (toString httpError)

        GetRequestedEntitlementsCompleted (Ok items) ->
            ( { model | requested = Just items }, Cmd.none )

        GetRequestedEntitlementsCompleted (Err httpError) ->
            Debug.crash (toString httpError)

        GetMetadataCompleted (Ok items) ->
            ( { model | metadata = Just items }, getAcceptedEntitlements )

        GetMetadataCompleted (Err httpError) ->
            Debug.crash (toString httpError)

        AcceptEntitlement ent ->
            ( model, acceptEntitlement ent )

        AcceptEntitlementCompleted (Ok ent) ->
            ( model, getAcceptedEntitlements )

        AcceptEntitlementCompleted (Err httpError) ->
            Debug.crash (toString httpError)

        DeclineEntitlement ent ->
            ( model, declineEntitlement ent )

        DeclineEntitlementCompleted (Ok ent) ->
            ( model, getAcceptedEntitlements )

        DeclineEntitlementCompleted (Err httpError) ->
            Debug.crash (toString httpError)

        AmendEntitlement ent ->
            ( model, amendEntitlement ent )

        AmendEntitlementCompleted (Ok ent) ->
            ( model, getAcceptedEntitlements )

        AmendEntitlementCompleted (Err httpError) ->
            Debug.crash (toString httpError)



--RPC


nodeURL : String
nodeURL =
    "http://localhost:8080"


getMetadata : Cmd Msg
getMetadata =
    let
        request =
            Http.get (nodeURL ++ "/data/meta") Decoders.decodeMetadata
    in
        Http.send GetMetadataCompleted request


getAcceptedEntitlements : Cmd Msg
getAcceptedEntitlements =
    let
        request =
            Http.get (nodeURL ++ "/entitlements/accepted/") Decoders.decodeEntitlements
    in
        Http.send GetAcceptedEntitlementsCompleted request


getRequestedEntitlements : Cmd Msg
getRequestedEntitlements =
    let
        request =
            Http.get (nodeURL ++ "/entitlements/requests") Decoders.decodeEntitlements
    in
        Http.send GetRequestedEntitlementsCompleted request


acceptEntitlement : Decoders.Entitlement -> Cmd Msg
acceptEntitlement ent =
    let
        request =
            Http.get (nodeURL ++ "/entitlements/requests/" ++ ent.uid ++ "/accept") Decoders.decodeEntitlement
    in
        Http.send AcceptEntitlementCompleted request


declineEntitlement : Decoders.Entitlement -> Cmd Msg
declineEntitlement ent =
    let
        request =
            Http.get (nodeURL ++ "/entitlements/requests/" ++ ent.uid ++ "/decline") Decoders.decodeEntitlement
    in
        Http.send DeclineEntitlementCompleted request


entitlementEncoder : Decoders.Entitlement -> Json.Encode.Value
entitlementEncoder ent =
    Json.Encode.object
        [ ( "subject", Json.Encode.string ent.subject )
        , ( "level", accessLevelEncoder ent.level )
        , ( "uid", Json.Encode.string ent.uid )
        , ( "status", Json.Encode.string ent.status )
        ]


accessLevelEncoder : Decoders.AccessLevel -> Json.Encode.Value
accessLevelEncoder level =
    case level of
        Decoders.OwnerOnly ->
            Json.Encode.string "owner-only"

        Decoders.CanDiscover ->
            Json.Encode.string "can-discover"

        Decoders.CanAccess ->
            Json.Encode.string "can-access"


amendEntitlement : Decoders.Entitlement -> Cmd Msg
amendEntitlement ent =
    let
        request =
            Http.post (nodeURL ++ "/entitlements/accepted/" ++ ent.uid) (Http.jsonBody (entitlementEncoder ent)) Decoders.decodeEntitlement
    in
        Http.send AmendEntitlementCompleted request



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
                    ]
                ]


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
            findEntitlement m.subject model.accepted

        requested =
            findEntitlement m.subject model.requested
    in
        div [] [ text (" current : "), (drawAccepted accepted), drawRequested (requested) ]


drawAccepted : Maybe Decoders.Entitlement -> Html Msg
drawAccepted ent =
    case ent of
        Nothing ->
            text ("entitlement not set")

        Just e ->
            div [] [ drawAccessLevel (e.level), text " ", drawAccessLevelSelector (e) ]


drawAccessLevelSelector : Decoders.Entitlement -> Html Msg
drawAccessLevelSelector ent =
    case ent.level of
        Decoders.OwnerOnly ->
            Html.span [] [ a [ onClick (AmendEntitlement { ent | level = Decoders.CanDiscover }), href "#" ] [ text ("make searchable") ] ]

        Decoders.CanDiscover ->
            Html.span []
                [ a [ onClick (AmendEntitlement { ent | level = Decoders.OwnerOnly }), href "#" ] [ text ("stop making available for search") ]
                , text (" ")
                , a [ onClick (AmendEntitlement { ent | level = Decoders.CanAccess }), href "#" ] [ text ("allow access to values") ]
                ]

        Decoders.CanAccess ->
            Html.span [] [ a [ onClick (AmendEntitlement { ent | level = Decoders.CanDiscover }), href "#" ] [ text ("remove access") ] ]


drawAccessLevel : Decoders.AccessLevel -> Html Msg
drawAccessLevel level =
    case level of
        Decoders.OwnerOnly ->
            text ("owner-only")

        Decoders.CanDiscover ->
            text ("can-discover")

        Decoders.CanAccess ->
            text ("can-access")


drawRequested : Maybe Decoders.Entitlement -> Html Msg
drawRequested ent =
    case ent of
        Nothing ->
            text ("")

        Just e ->
            div []
                [ text (" requested : ")
                , drawAccessLevel (e.level)
                , a [ onClick (AcceptEntitlement e), href "#" ] [ text ("accept") ]
                , text (" ")
                , a [ onClick (DeclineEntitlement e), href "#" ] [ text ("decline") ]
                ]


findEntitlement : String -> Maybe Decoders.Entitlements -> Maybe Decoders.Entitlement
findEntitlement key entitlements =
    case entitlements of
        Nothing ->
            Nothing

        Just ents ->
            List.Extra.find (\e -> e.subject == key) ents
