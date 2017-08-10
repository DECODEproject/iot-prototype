module Node exposing (..)

import Http
import Html exposing (..)
import Html.Attributes exposing (..)
import Html.Events exposing (onClick)
import List.Extra exposing (..)
import Json.Encode exposing (..)
import Bootstrap.CDN as CDN
import Bootstrap.Grid as Grid
import Bootstrap.Tab as Tab
import Bootstrap.Table as Table
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
    , tabState : Tab.State
    }


initialModel : Model
initialModel =
    { accepted = Nothing
    , requested = Nothing
    , metadata = Nothing
    , tabState = Tab.initialState
    }


init : ( Model, Cmd Msg )
init =
    ( initialModel, getMetadata )


type Msg
    = NoOp
    | Refresh
    | GetMetadataCompleted (Result Http.Error Decoders.Metadata)
    | GetAcceptedEntitlementsCompleted (Result Http.Error Decoders.Entitlements)
    | GetRequestedEntitlementsCompleted (Result Http.Error Decoders.Entitlements)
    | AcceptEntitlement Decoders.Entitlement
    | AcceptEntitlementCompleted (Result Http.Error Decoders.Entitlement)
    | DeclineEntitlement Decoders.Entitlement
    | DeclineEntitlementCompleted (Result Http.Error Decoders.Entitlement)
    | AmendEntitlement Decoders.Entitlement
    | AmendEntitlementCompleted (Result Http.Error Decoders.Entitlement)
    | TabMsg Tab.State


update : Msg -> Model -> ( Model, Cmd Msg )
update msg model =
    case msg of
        NoOp ->
            ( model, Cmd.none )

        Refresh ->
            ( model, getMetadata )

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

        TabMsg state ->
            ( { model | tabState = state }, Cmd.none )



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
    div []
        [ CDN.stylesheet
        , Tab.config TabMsg
            |> Tab.items
                [ Tab.item
                    { id = "tabItem1"
                    , link = Tab.link [] [ text "Devices" ]
                    , pane = deviceTab model
                    }
                , Tab.item
                    { id = "tabItem2"
                    , link = Tab.link [] [ text "Entitlements" ]
                    , pane = entitlementsTab model
                    }
                ]
            |> Tab.view model.tabState
        ]


deviceTab : Model -> Tab.Pane msg
deviceTab model =
    Tab.pane [ Html.Attributes.class "mt-3" ]
        [ h4 [] [ text "Devices" ]
        , p [] [ text "This is the page where you can add, remove devices." ]
        , text "coming soon"
        ]


entitlementsTab : Model -> Tab.Pane Msg
entitlementsTab model =
    Tab.pane [ Html.Attributes.class "mt-3" ]
        [ h4 [] [ text "Entitlements" ]
        , p [] [ text "This page is where you can edit, view, accept and reject entitlements to your data." ]
        , button [ onClick Refresh ] [ text "Refresh" ]
        , div [] []
        , entitlementsTable model
        ]


entitlementsTable : Model -> Html Msg
entitlementsTable model =
    Table.simpleTable
        ( Table.simpleThead
            [ Table.th [] [ text "Data" ]
            , Table.th [] [ text "Current" ]
            , Table.th [] [ text "Proposed" ]
            ]
        , Table.tbody [] <| drawMetadata model
        )


drawMetadata : Model -> List (Table.Row Msg)
drawMetadata model =
    case model.metadata of
        Nothing ->
            -- there must be a better way than this!
            List.map (\m -> Table.tr [] [ Table.td [] [ text ("no data exists.") ] ]) [ 1 ]

        Just e ->
            List.map (\m -> drawMetadataItem m model) e


drawMetadataItem : Decoders.MetadataItem -> Model -> Table.Row Msg
drawMetadataItem m model =
    let
        accepted =
            findEntitlement m.subject model.accepted

        requested =
            findEntitlement m.subject model.requested
    in
        Table.tr [] [ Table.td [] [ text (m.description) ], Table.td [] [ drawAccepted (accepted) ], Table.td [] [ drawRequested (requested) ] ]


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
            Html.span [] [ drawAccessLevel (e.level), text " ", drawAccessLevelSelector (e) ]


drawAccessLevelSelector : Decoders.Entitlement -> Html Msg
drawAccessLevelSelector ent =
    case ent.level of
        Decoders.OwnerOnly ->
            div [] [ a [ onClick (AmendEntitlement { ent | level = Decoders.CanDiscover }), href "#" ] [ text ("make data searchable") ] ]

        Decoders.CanDiscover ->
            div []
                [ a [ onClick (AmendEntitlement { ent | level = Decoders.OwnerOnly }), href "#" ] [ text ("stop making available for search") ]
                , text (" ")
                , a [ onClick (AmendEntitlement { ent | level = Decoders.CanAccess }), href "#" ] [ text ("make data accessible") ]
                ]

        Decoders.CanAccess ->
            div [] [ a [ onClick (AmendEntitlement { ent | level = Decoders.CanDiscover }), href "#" ] [ text ("remove access") ] ]


drawAccessLevel : Decoders.AccessLevel -> Html Msg
drawAccessLevel level =
    case level of
        Decoders.OwnerOnly ->
            text ("Only the owner (you) can see the data")

        Decoders.CanDiscover ->
            text ("Anyone can discover the data")

        Decoders.CanAccess ->
            text ("Anyone can access the data")


drawRequested : Maybe Decoders.Entitlement -> Html Msg
drawRequested ent =
    case ent of
        Nothing ->
            text ("")

        Just e ->
            div []
                [ drawAccessLevel (e.level)
                , div []
                    [ a [ onClick (AcceptEntitlement e), href "#" ] [ text ("accept") ]
                    , text (" ")
                    , a [ onClick (DeclineEntitlement e), href "#" ] [ text ("decline") ]
                    ]
                ]


findEntitlement : String -> Maybe Decoders.Entitlements -> Maybe Decoders.Entitlement
findEntitlement key entitlements =
    case entitlements of
        Nothing ->
            Nothing

        Just ents ->
            List.Extra.find (\e -> e.subject == key) ents
