module MetadataClient exposing (..)

import Json.Decode exposing (Decoder, string, int, float, dict, list, bool, map, value, decodeValue, decodeString, lazy, succeed, fail, andThen)
import Json.Decode.Pipeline exposing (decode, required, optional, hardcoded)
import Dict exposing (Dict)

type alias Items =
    List Item


decodeItems : Json.Decode.Decoder (List Item)
decodeItems =
    Json.Decode.list decodeItem


type alias Item =
    { key : String
    , location : Location
    , sample : String
    , tags : List String
    , uid : String
    }


decodeItem : Json.Decode.Decoder Item
decodeItem =
    Json.Decode.map5 Item
        (Json.Decode.field "key" Json.Decode.string)
        (Json.Decode.field "location" decodeLocation)
        (Json.Decode.field "sample" Json.Decode.string)
        (Json.Decode.field "tags" (Json.Decode.list Json.Decode.string))
        (Json.Decode.field "uid" Json.Decode.string)


type alias Location =
    { ipAddress : String
    , ipPort : Int
    , uid : String
    }


decodeLocation : Json.Decode.Decoder Location
decodeLocation =
    Json.Decode.map3 Location
        (Json.Decode.field "ip-address" Json.Decode.string)
        (Json.Decode.field "port" Json.Decode.int)
        (Json.Decode.field "uid" Json.Decode.string)
