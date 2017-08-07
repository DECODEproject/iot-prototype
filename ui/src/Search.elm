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



-- MODEL


type alias Model =
    { all : Maybe Decoders.Items
    , filter : Maybe String
    , currentGraph : Maybe Decoders.DataResponse
    , currentItem : Maybe Decoders.Item
    }


initialModel : Model
initialModel =
    { all = Nothing
    , filter = Nothing
    , currentGraph = Nothing
    , currentItem = Nothing
    }


init : ( Model, Cmd Msg )
init =
    ( initialModel, Cmd.none )



-- UPDATE


type Msg
    = NoOp
    | RefreshMetadata
    | RefreshMetadataCompleted (Result Http.Error Decoders.Items)
    | ShowLocations String
    | ViewGraph Decoders.Item
    | ViewGraphCompleted (Result Http.Error Decoders.DataResponse)
    | RequestAccess Decoders.Item
    | RequestAccessCompleted (Result Http.Error Decoders.Entitlement)


update : Msg -> Model -> ( Model, Cmd Msg )
update msg model =
    case msg of
        NoOp ->
            ( model, Cmd.none )

        RefreshMetadata ->
            ( model, getAllMetadata )

        RefreshMetadataCompleted (Ok items) ->
            ( { model | all = Just items }, Cmd.none )

        RefreshMetadataCompleted (Err httpError) ->
            let
                _ =
                    Debug.log "RefreshMetadata error" httpError
            in
                ( model, Cmd.none )

        ShowLocations tag ->
            case model.all of
                Nothing ->
                    ( model, Cmd.none )

                Just r ->
                    ( { model | filter = Just tag, currentGraph = Nothing }, Cmd.none )

        ViewGraph item ->
            ( { model | currentGraph = Nothing, currentItem = Just item }, getTimeSeriesData item )

        ViewGraphCompleted (Err (Http.BadStatus response)) ->
            let
                _ =
                    Debug.log "xxx" response

                items =
                    updateRight model.all model.currentItem Decoders.RequestAccess
            in
                ( { model | all = items }, Cmd.none )

        ViewGraphCompleted (Err httpError) ->
            let
                _ =
                    Debug.log "ViewGraph error" httpError
            in
                ( model, Cmd.none )

        ViewGraphCompleted (Ok items) ->
            ( { model | currentGraph = Just items }, unsafeDrawGraph (prepareGraphData (items.data)) )

        RequestAccess item ->
            -- make entitlement request
            -- update model
            ( model, requestAccess item )

        RequestAccessCompleted (Err httpError) ->
            let
                _ =
                    Debug.log "RequestAccess error" httpError
            in
                ( model, Cmd.none )

        RequestAccessCompleted (Ok items) ->
            ( model, Cmd.none )


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


getAllMetadata : Cmd Msg
getAllMetadata =
    let
        url =
            "http://localhost:8081/catalog/items/"

        request =
            Http.get url Decoders.decodeItems
    in
        Http.send RefreshMetadataCompleted request


getTimeSeriesEncoder : String -> Json.Encode.Value
getTimeSeriesEncoder key =
    Json.Encode.object [ ( "key", Json.Encode.string key ) ]


getTimeSeriesData : Decoders.Item -> Cmd Msg
getTimeSeriesData item =
    let
        url =
            "http://localhost:8080/data/"

        request =
            Http.post url (Http.jsonBody (getTimeSeriesEncoder item.key)) Decoders.decodeDataResponse
    in
        Http.send ViewGraphCompleted request


entitlementRequestEncoder : Decoders.Item -> Json.Encode.Value
entitlementRequestEncoder item =
    Json.Encode.object
        [ ( "level", Json.Encode.string "can-read" )
        , ( "subject", Json.Encode.string item.key )
        ]


requestAccess : Decoders.Item -> Cmd Msg
requestAccess item =
    let
        url =
            "http://localhost:8080/entitlements/requests/"

        request =
            Http.request
                { method = "PUT"
                , headers = []
                , url = url
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
            button [ onClick RefreshMetadata ] [ text "refresh" ]

        Just d ->
            div []
                [ div [] [ text "Metadata" ]
                , drawData d
                , drawFiltered model.filter d
                , div [] [ button [ onClick RefreshMetadata ] [ text "refresh" ] ]

                --                , div [] [ text (toString model) ]
                ]


drawData : Decoders.Items -> Html Msg
drawData items =
    div [] <| List.map (\x -> div [] [ a [ onClick (ShowLocations x), href "#" ] [ text (x) ] ]) (uniqueTags items)


drawFiltered : Maybe String -> Decoders.Items -> Html Msg
drawFiltered tag items =
    case tag of
        Nothing ->
            text ("")

        Just t ->
            let
                filtered =
                    items
            in
                div [] <| List.map (\item -> div [] [ text item.key, text (toString item.location), drawViewerWidget (item) ]) filtered


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


updateRight : Maybe Decoders.Items -> Maybe Decoders.Item -> Decoders.Right -> Maybe Decoders.Items
updateRight items item right =
    case items of
        Nothing ->
            items

        Just all ->
            case item of
                Nothing ->
                    items

                Just i ->
                    let
                        location1 =
                            i.location

                        location2 =
                            { location1 | right = right }
                    in
                        Just (List.Extra.updateIf (\n -> n.uid == i.uid) (\t -> { t | location = location2 }) all)


uniqueLocations : Decoders.Items -> List String
uniqueLocations items =
    List.map (\x -> x.location.uid) items
        |> List.Extra.unique


uniqueTags : Decoders.Items -> List String
uniqueTags items =
    List.concatMap (\x -> x.tags) items
        |> List.Extra.unique


filterByTag : String -> Decoders.Items -> Maybe Decoders.Items
filterByTag tag data =
    let
        r =
            List.filter (\x -> List.member tag x.tags) data
    in
        case r of
            [] ->
                Nothing

            _ ->
                Just r
