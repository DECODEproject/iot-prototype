module Main exposing (..)

import Http
import MetadataClient

import Html exposing (..)
import Html.Attributes exposing (..)
import Html.Events exposing (onClick)


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
    {
        all : Maybe MetadataClient.Items
}


initialModel : Model
initialModel =
    {
        all = Nothing
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
                    ( {model | all = Just items }, Cmd.none )


getAllMetadata :  Cmd Msg
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
    div []
        [ text "Nodes"
        , text "Data"
        , button [ onClick RefreshMetadata] [text "refresh"]
        , text (toString model)
    ]
