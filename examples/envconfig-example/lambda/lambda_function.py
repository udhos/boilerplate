import json
import base64

secret = 'secret'

def lambda_handler(event, context):
    
    event_str = json.dumps(event)
    
    print("lambda_handler: event:", event_str)
    
    headers = event.get('headers')
    if headers is None:
        #
        # lambda invoked directly
        #
        return handle_parameter(event)
        
    #
    # lambda invoked from function url
    #
    # curl -H 'content-type: application/json' -H 'authorization: Bearer secret' -d '{"parameter":"mongodb"}' https://ttt.lambda-url.us-east-1.on.aws/
    #
    
    auth = headers.get('authorization')
    if auth is None:
        print("missing header authorization")
        return forbidden()
    
    fields = auth.split(None, 1)
    if len(fields) < 2:
        print("missing token in header authorization")
        return forbidden()
    
    token = fields[1]
    if token != secret:
        print("invalid token in header authorization")
        return forbidden()
    
    # If the content type of the request is binary, the body is base64-encoded.
    body = event.get('body')
    try:
        d = decode(body)
        print("decode: ", d)
        body = d
    except Exception:
        pass
    
    print("body STR:", body)
    body_obj = json.loads(body)
    print("body OBJ:", body_obj)
    return handle_parameter(body_obj)

def decode(base64_message):
    base64_bytes = base64_message.encode('utf-8')
    message_bytes = base64.b64decode(base64_bytes)
    message = message_bytes.decode('utf-8')
    return message

def handle_parameter(request):
    print("handle_parameter: request:", request)
    param = request.get('parameter')
    print("handle_parameter: parameter:", param)
    if param is None:
        print("missing parameter")
        return bad_request()
    if param == 'mongodb':
        return return_body('{"uri": "mongodb://localhost:27017/?retryWrites=false"}')
    print("handle_parameter: parameter not found: request:", request)
    return bad_request()
    

def forbidden():
    return {
        "statusCode": 403,
        "body": "forbidden"
    }

def bad_request():
    return {
        "statusCode": 400,
        "body": "bad request"
    }

def return_body(body):
    return {
        "statusCode": 200,
        "body": body
    }
