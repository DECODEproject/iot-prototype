module Main exposing (..)

import Http
import Html exposing (..)
import Html.Attributes exposing (..)
import Html.Events exposing (onClick)
import List.Extra exposing (unique)
import MetadataClient


main : Program Never Model Msg
main =
    Html.program
        { init = init
        , view = view
        , update = update
        , subscriptions = subscriptions
        }



-- MODEL


type alias Model =
    { all : Maybe MetadataClient.Items
    , filtered : Maybe MetadataClient.Items
    }


initialModel : Model
initialModel =
    { all = Nothing
    , filtered = Nothing
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
                    ( { model | filtered = (filterByTag tag r) }, Cmd.none )

        ViewGraph key location ->
            ( model, Cmd.none )


getAllMetadata : Cmd Msg
getAllMetadata =
    let
        url =
            "http://localhost:8081/catalog/items/"

        request =
            Http.get url MetadataClient.decodeItems
    in
        Http.send RefreshMetadataCompleted request



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
                , div [] [ text (toString model) ]
                ]



--drawNodes : MetadataClient.Items -> Html Msg
--drawNodes items =
--    div [] <| List.map (\x -> div [] [ text (x) ]) (uniqueLocations items)


drawData : MetadataClient.Items -> Html Msg
drawData items =
    div [] <| List.map (\x -> div [] [ a [ onClick (ShowLocations x), href "#" ] [ text (x) ] ]) (uniqueTags items)



--drawKeys : MetadataClient.Items -> Html Msg
--drawKeys items =
--    div [] <| List.map (\x -> div [] [ text (x.key) ]) items


drawFiltered : Maybe MetadataClient.Items -> Html Msg
drawFiltered items =
    case items of
        Nothing ->
            text ("")

        Just r ->
            div [] <| List.map (\x -> div [] [ text x.key, text (toString (x.location)), a [ onClick (ViewGraph x.key x.location), href "#" ] [ text "view" ] ]) r


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
