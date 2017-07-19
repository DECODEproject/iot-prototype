module NodeClient exposing(..)

import Json.Decode exposing (Decoder, string, int, float, dict, list, bool, map, value, decodeValue, decodeString, lazy, succeed, fail, andThen)
import Json.Decode.Pipeline exposing (decode, required, optional, hardcoded)
import Dict exposing (Dict)


type alias DataResponse =
    {data : List DataItem
    }


decodeDataResponse : Decoder DataResponse
decodeDataResponse =
    Json.Decode.map DataResponse
        (Json.Decode.field "data" (Json.Decode.list decodeDataItem))

type alias DataItem =
    { value :JsVal
    ,timeStamp :String
    }

decodeDataItem : Decoder DataItem
decodeDataItem =
    Json.Decode.map2 DataItem
        (Json.Decode.field "value" jsValDecoder)
        (Json.Decode.field "ts" Json.Decode.string)

type JsVal
  = JsString String
  | JsInt Int
  | JsFloat Float
  | JsArray (List JsVal)
  | JsObject (Dict String JsVal)
  | JsNull

jsValDecoder : Decoder JsVal
jsValDecoder =
      Json.Decode.oneOf
        [ Json.Decode.map JsString Json.Decode.string
        , Json.Decode.map JsInt Json.Decode.int
        , Json.Decode.map JsFloat Json.Decode.float
        , Json.Decode.list (Json.Decode.lazy (\_ -> jsValDecoder)) |> Json.Decode.map JsArray
        , Json.Decode.dict (Json.Decode.lazy (\_ -> jsValDecoder)) |> Json.Decode.map JsObject
        , Json.Decode.null JsNull
        ]
