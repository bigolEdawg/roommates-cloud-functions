from flask import Flask, jsonify, request
from datetime import datetime
import uuid

app = Flask(__name__)

@app.route('/api/user/groups', methods=['GET'])
def get_user_groups():
    # Verify the token (in production, you'd actually verify Firebase token)
    # auth_header = request.headers.get('Authorization')
    # if not auth_header or 'Bearer' not in auth_header:
    #     return jsonify({"error": "Unauthorized"}), 401
    
    # Extract user ID from token (for now, just use a placeholder)
    # In production: verify Firebase token and extract UID
    user_id = "test-user-123"
    
    # Return data in the format Swift expects
    groups = [
        {
            "id": str(uuid.uuid4()),
            "name": "Roommates - San Francisco",
            "ownerId": user_id,
            "createdAt": datetime.now().isoformat(),
            "members": [user_id, "user-456", "user-789"]
        },
        {
            "id": str(uuid.uuid4()),
            "name": "College Friends",
            "ownerId": user_id,
            "createdAt": datetime.now().isoformat(),
            "members": [user_id, "user-abc"]
        }
    ]
    
    return jsonify(groups)

if __name__ == "__main__":
    app.run(debug=True)

# from flask import Flask, jsonify

# app = Flask(__name__)

# data = [ 
#         {
#             "group-id" : "1",
#             "group-name" : "first group",
#             "user-id" : ""
#         },
#         {
#             "group-id" : "2",
#             "group-name" : "second group",
#             "user-id" : ""
#         }
#     ]

# @app.route('/api/user/groups', methods=['GET'])
# def testing():
 
#     return jsonify(data)

# if __name__ == "__main__":
#     app.run(debug=True)



