port module Main exposing (..)

import Date
import Http
import Html exposing (..)
import Html.Attributes exposing (..)
import Html.Events exposing (onClick)
import List.Extra exposing (unique)
import MetadataClient
import NodeClient

import Json.Encode exposing(..)

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
    { all : Maybe MetadataClient.Items
    , filtered : Maybe MetadataClient.Items
    , currentGraph : Maybe NodeClient.DataResponse
    }


initialModel : Model
initialModel =
    { all = Nothing
    , filtered = Nothing
    , currentGraph = Nothing
    }


init : ( Model, Cmd Msg )
init =
    ( initialModel, Cmd.none )



-- UPDATE


type Msg
    = NoOp
    | RefreshMetadata
    | RefreshMetadataCompleted (Result Http.Error MetadataClient.Items)
    | ShowLocations String
    | ViewGraph String MetadataClient.Location
    | ViewGraphCompleted (Result Http.Error NodeClient.DataResponse)


update : Msg -> Model -> ( Model, Cmd Msg )
update msg model =
    case msg of
        NoOp ->
            ( model, Cmd.none )

        RefreshMetadata ->
            ( model, getAllMetadata )

        RefreshMetadataCompleted result ->
            case result of
                Err httpError ->
                    let
                        _ =
                            Debug.log "RefreshMetadata error" httpError
                    in
                        ( model, Cmd.none )

                Ok items ->
                    ( { model | all = Just items }, Cmd.none )

        ShowLocations tag ->
            case model.all of
                Nothing ->
                    ( model, Cmd.none )

                Just r ->
                    ( { model | filtered = (filterByTag tag r), currentGraph = Nothing }, Cmd.none )

        ViewGraph key location ->
            ( {model | currentGraph = Nothing }, getTimeSeriesData key )
        ViewGraphCompleted result ->
            case result of
                Err httpError ->
                    let
                        _ =
                            Debug.log "ViewGraph error" httpError
                    in
                        ( model, Cmd.none )

                Ok items ->
                    ( { model | currentGraph = Just items }, unsafeDrawGraph( prepareGraphData(items.data) ))


type alias FloatDataItem =
    {
        value : Float,
        date : String
    }

prepareGraphData : List NodeClient.DataItem -> List FloatDataItem
prepareGraphData items =
    List.filterMap (\item -> case item.value of
            NodeClient.JsFloat f ->  Just (FloatDataItem f item.timeStamp)
            _ -> Nothing
        ) items


getAllMetadata : Cmd Msg
getAllMetadata =
    let
        url =
            "http://localhost:8081/catalog/items/"

        request =
            Http.get url MetadataClient.decodeItems
    in
        Http.send RefreshMetadataCompleted request


getTimeSeriesEncoder : String -> Json.Encode.Value
getTimeSeriesEncoder key =
    Json.Encode.object [ ("key", Json.Encode.string key) ]

getTimeSeriesData : String -> Cmd Msg
getTimeSeriesData key =
    let
        url =
            "http://localhost:8080/data/"

        request =
            Http.post url (Http.jsonBody (getTimeSeriesEncoder key)) NodeClient.decodeDataResponse
    in
        Http.send ViewGraphCompleted request


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
                , drawFiltered model.filtered
                , div [] [ button [ onClick RefreshMetadata ] [ text "refresh" ] ]
--                , div [] [ text (toString model) ]
                ]


drawData : MetadataClient.Items -> Html Msg
drawData items =
    div [] <| List.map (\x -> div [] [ a [ onClick (ShowLocations x), href "#" ] [ text (x) ] ]) (uniqueTags items)

drawFiltered : Maybe MetadataClient.Items -> Html Msg
drawFiltered items =
    case items of
        Nothing ->
            text ("")

        Just items ->
            div [] <| List.map (\item -> div [] [ text item.key, text (toString item.location), drawViewerWidget(item) ]) items


drawViewerWidget : MetadataClient.Item -> Html Msg
drawViewerWidget item =
    a [ onClick (ViewGraph item.key item.location), href "#" ] [ text "view" ]

-- RPC
-- TODO :  move to MetadataClient
uniqueLocations : MetadataClient.Items -> List String
uniqueLocations items =
    List.map (\x -> x.location.uid) items
        |> List.Extra.unique


uniqueTags : MetadataClient.Items -> List String
uniqueTags items =
    List.concatMap (\x -> x.tags) items
        |> List.Extra.unique


filterByTag : String -> MetadataClient.Items -> Maybe MetadataClient.Items
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
