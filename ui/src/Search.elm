port module Search exposing (..)

import Date
import Http
import Html exposing (..)
import Html.Attributes exposing (..)
import Html.Events exposing (onClick)
import List.Extra exposing (unique)
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


port unsafeDrawGraph : List FloatDataItem -> Cmd msg


port unsafeClearGraph : String -> Cmd msg



-- MODEL


type alias Model =
    { all : Maybe Decoders.Items
    , filter : Maybe String
    }


initialModel : Model
initialModel =
    { all = Nothing
    , filter = Nothing
    }


init : ( Model, Cmd Msg )
init =
    ( initialModel, getAllMetadata )



-- UPDATE


type Msg
    = NoOp
    | RefreshMetadata
    | RefreshMetadataCompleted (Result Http.Error Decoders.Items)
    | ShowLocations String
    | ViewGraph Decoders.Item
    | ViewGraphCompleted Decoders.Item (Result Http.Error Decoders.DataResponse)
    | RequestAccess Decoders.Item
    | RequestAccessCompleted (Result Http.Error Decoders.Entitlement)


update : Msg -> Model -> ( Model, Cmd Msg )
update msg model =
    case msg of
        NoOp ->
            ( model, Cmd.none )

        RefreshMetadata ->
            ( { model | filter = Nothing }, getAllMetadata )

        RefreshMetadataCompleted (Ok items) ->
            ( { model | all = Just items }, unsafeClearGraph "only-for-the-compiler" )

        RefreshMetadataCompleted (Err httpError) ->
            Debug.crash (toString httpError)

        ShowLocations tag ->
            case model.all of
                Nothing ->
                    ( model, Cmd.none )

                Just r ->
                    ( { model | filter = Just tag }, Cmd.none )

        ViewGraph item ->
            ( model, getTimeSeriesData item )

        ViewGraphCompleted requested (Ok graphData) ->
            let
                items =
                    updateRight model.all requested Decoders.Unknown
            in
                ( { model | all = items }, unsafeDrawGraph (prepareGraphData (graphData.data)) )

        ViewGraphCompleted requested (Err (Http.BadStatus response)) ->
            let
                items =
                    updateRight model.all requested Decoders.RequestAccess
            in
                ( { model | all = items }, Cmd.none )

        ViewGraphCompleted requested (Err httpError) ->
            Debug.crash (toString httpError)

        RequestAccess item ->
            let
                items =
                    updateRight model.all item Decoders.Requesting
            in
                ( { model | all = items }, requestAccess item )

        RequestAccessCompleted (Ok items) ->
            ( model, Cmd.none )

        RequestAccessCompleted (Err httpError) ->
            Debug.crash (toString httpError)


type alias FloatDataItem =
    { value : Float
    , date : String
    }


prepareGraphData : List Decoders.DataItem -> List FloatDataItem
prepareGraphData items =
    List.filterMap
        (\item ->
            case item.value of
                Decoders.JsFloat f ->
                    Just (FloatDataItem f item.timeStamp)

                Decoders.JsInt i ->
                    Just (FloatDataItem (toFloat i) item.timeStamp)

                _ ->
                    Nothing
        )
        items



-- RPC


nodeURLFromLocation : Decoders.Location -> String
nodeURLFromLocation location =
    location.scheme ++ "://" ++ location.ipAddress ++ ":" ++ toString (location.ipPort)


metadataURL : String
metadataURL =
    "http://localhost:8081"


getAllMetadata : Cmd Msg
getAllMetadata =
    let
        request =
            Http.get (metadataURL ++ "/catalog/items/") Decoders.decodeItems
    in
        Http.send RefreshMetadataCompleted request


getTimeSeriesEncoder : String -> Json.Encode.Value
getTimeSeriesEncoder key =
    Json.Encode.object [ ( "key", Json.Encode.string key ) ]


getTimeSeriesData : Decoders.Item -> Cmd Msg
getTimeSeriesData item =
    let
        request =
            Http.post (nodeURLFromLocation (item.location) ++ "/data/") (Http.jsonBody (getTimeSeriesEncoder item.key)) Decoders.decodeDataResponse
    in
        Http.send (ViewGraphCompleted item) request


entitlementRequestEncoder : Decoders.Item -> Json.Encode.Value
entitlementRequestEncoder item =
    Json.Encode.object
        [ ( "level", Json.Encode.string "can-access" )
        , ( "subject", Json.Encode.string item.key )
        ]


requestAccess : Decoders.Item -> Cmd Msg
requestAccess item =
    let
        request =
            Http.request
                { method = "PUT"
                , headers = []
                , url = nodeURLFromLocation (item.location) ++ "/entitlements/requests/"
                , body = Http.jsonBody (entitlementRequestEncoder item)
                , expect = Http.expectJson Decoders.decodeEntitlement
                , timeout = Nothing
                , withCredentials = False
                }
    in
        Http.send RequestAccessCompleted request



-- SUBSCRIPTIONS


subscriptions : Model -> Sub Msg
subscriptions model =
    Sub.none



-- VIEW


view : Model -> Html Msg
view model =
    case model.all of
        Nothing ->
            drawNoMetadata

        Just d ->
            case d of
                [] ->
                    drawNoMetadata

                _ ->
                    div []
                        [ div [] [ text "Metadata" ]
                        , div [] [ drawFiltered model.filter d ]
                        ]


drawNoMetadata : Html Msg
drawNoMetadata =
    div []
        [ div [] [ text "no metadata available." ]
        , button [ onClick RefreshMetadata ] [ text "refresh" ]
        ]


drawFiltered : Maybe String -> Decoders.Items -> Html Msg
drawFiltered tag items =
    case tag of
        Nothing ->
            div [] <| List.map (\x -> div [] [ a [ onClick (ShowLocations x), href "#" ] [ text (x) ] ]) (uniqueTags items)

        Just t ->
            let
                filtered =
                    filterByTag t items
            in
                div []
                    [ div [] [ text (t) ]
                    , div [] <| List.map (\item -> div [] [ text item.key, text " ", drawViewerWidget (item) ]) filtered
                    , div [] [ button [ onClick RefreshMetadata ] [ text "new search" ] ]
                    ]


drawViewerWidget : Decoders.Item -> Html Msg
drawViewerWidget item =
    case item.location.right of
        Decoders.Unknown ->
            a [ onClick (ViewGraph item), href "#" ] [ text "view" ]

        Decoders.RequestAccess ->
            a [ onClick (RequestAccess item), href "#" ] [ text "request access" ]

        Decoders.Requesting ->
            a [ onClick (ViewGraph item), href "#" ] [ text "request in progress, try again" ]



-- Helpers


updateRight : Maybe Decoders.Items -> Decoders.Item -> Decoders.Right -> Maybe Decoders.Items
updateRight items item right =
    case items of
        Nothing ->
            items

        Just all ->
            let
                location1 =
                    item.location

                location2 =
                    { location1 | right = right }
            in
                Just (List.Extra.updateIf (\n -> n.uid == item.uid) (\t -> { t | location = location2 }) all)


uniqueLocations : Decoders.Items -> List String
uniqueLocations items =
    List.map (\x -> x.location.uid) items
        |> List.Extra.unique


uniqueTags : Decoders.Items -> List String
uniqueTags items =
    List.concatMap (\x -> x.tags) items
        |> List.Extra.unique


filterByTag : String -> Decoders.Items -> Decoders.Items
filterByTag tag data =
    List.filter (\x -> List.member tag x.tags) data
