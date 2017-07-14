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
    }


initialModel : Model
initialModel =
    { all = Nothing
    }


init : ( Model, Cmd Msg )
init =
    ( initialModel, Cmd.none )



-- UPDATE


type Msg
    = NoOp
    | RefreshMetadata
    | RefreshMetadataCompleted (Result Http.Error MetadataClient.Items)


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
                [ div [] [ text "Nodes" ]
                , drawNodes d
                , div [] [ text "Data" ]
                , drawData d
                , div [] [ text "Keys" ]
                , drawKeys d
                , div [] [ button [ onClick RefreshMetadata ] [ text "refresh" ] ]

                --, div [] [text (toString model)]
                ]


drawNodes : MetadataClient.Items -> Html Msg
drawNodes items =
    div [] <| List.map (\x -> text (x)) (uniqueLocations items)


drawData : MetadataClient.Items -> Html Msg
drawData items =
    div [] <| List.map (\x -> text (x)) (uniqueTags items)


drawKeys : MetadataClient.Items -> Html Msg
drawKeys items =
    div [] <| List.map (\x -> text (x.key)) items


uniqueLocations : MetadataClient.Items -> List String
uniqueLocations items =
    List.map (\x -> x.location.uid) items
        |> List.Extra.unique


uniqueTags : MetadataClient.Items -> List String
uniqueTags items =
    List.concatMap (\x -> x.tags) items
        |> List.Extra.unique
