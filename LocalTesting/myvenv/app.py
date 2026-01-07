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
            "id": "1",
            "name": "Roommates - San Francisco",
            "ownerId": user_id,
            "createdAt": datetime.now().isoformat(),
            "members": [user_id, "user-456", "user-789"]
        },
        {
            "id": "2",
            "name": "College Friends",
            "ownerId": user_id,
            "createdAt": datetime.now().isoformat(),
            "members": [user_id, "user-abc"]
        }
    ]
    
    return jsonify(groups)

chores = {
    "1": [
        {
            "chore_assignee": "u_123",
            "chore_details": "Blue bin + garbage",
            "chore_due_date": "2025-10-01T02:00:00Z",
            "chore_frequency": "weekly",
            "chore_name": "Take out trash",
            "group_id": "1"
        },
        {
            "chore_assignee": "u_456",
            "chore_details": "Vacuum all common areas",
            "chore_due_date": "2025-10-03T18:00:00Z",
            "chore_frequency": "bi-weekly",
            "chore_name": "Vacuum floors",
            "group_id": "1"
        },
        {
            "chore_assignee": "u_789",
            "chore_details": "Clean microwave, stove, and countertops",
            "chore_due_date": "2025-10-02T20:00:00Z",
            "chore_frequency": "weekly",
            "chore_name": "Clean kitchen",
            "group_id": "1"
        }
    ],
    "2": [
        {
            "chore_assignee": "u_abc",
            "chore_details": "Wipe down mirrors and sinks",
            "chore_due_date": "2025-10-04T12:00:00Z",
            "chore_frequency": "weekly",
            "chore_name": "Clean bathroom",
            "group_id": "2"
        },
        {
            "chore_assignee": "u_def",
            "chore_details": "Separate glass, plastic, and paper",
            "chore_due_date": "2025-10-05T15:00:00Z",
            "chore_frequency": "monthly",
            "chore_name": "Sort recycling",
            "group_id": "2"
        }
    ],
    "3": [
        {
            "chore_assignee": "u_xyz",
            "chore_details": "Water all indoor plants",
            "chore_due_date": "2025-10-06T10:00:00Z",
            "chore_frequency": "weekly",
            "chore_name": "Water plants",
            "group_id": "3"
        },
        {
            "chore_assignee": "u_123",
            "chore_details": "Collect mail from mailbox",
            "chore_due_date": "2025-10-07T17:00:00Z",
            "chore_frequency": "daily",
            "chore_name": "Check mail",
            "group_id": "3"
        },
        {
            "chore_assignee": "u_456",
            "chore_details": "Wipe all windows inside",
            "chore_due_date": "2025-10-08T14:00:00Z",
            "chore_frequency": "monthly",
            "chore_name": "Clean windows",
            "group_id": "3"
        },
        {
            "chore_assignee": "u_789",
            "chore_details": "Take bins to curb on Tuesday night",
            "chore_due_date": "2025-10-09T19:00:00Z",
            "chore_frequency": "weekly",
            "chore_name": "Take out recycling",
            "group_id": "3"
        }
    ]
}

# let chore_assignee: String
# let chore_details: String
# let chore_due_date: String
# let chore_frequency: String
# let chore_name: String
# let group_id: String

@app.route('/api/group/chore', methods=['GET'])
def get_group_chore():
    group_id = request.args.get("group-id")

    chore_list = chores[group_id]


    return jsonify(chore_list)


members = {
    "1": [
        {
            "id": "1a",
            "added_by": "admin",
            "joined_at": "2024-01-15T10:30:00Z",
            "user_id": "user_123",
            "user_name": "Alex Johnson",
            "groupId": "1"
        },
        {
            "id": "1b",
            "added_by": "user_123",
            "joined_at": "2024-01-20T14:45:00Z",
            "user_id": "user_456",
            "user_name": "Sam Wilson",
            "groupId": "1"
        },
        {
            "id": "1c",
            "added_by": "user_123",
            "joined_at": "2024-02-05T09:15:00Z",
            "user_id": "user_789",
            "user_name": "Taylor Smith",
            "groupId": "1"
        },
        {
            "id": "1d",
            "added_by": "user_456",
            "joined_at": "2024-02-10T16:20:00Z",
            "user_id": "user_101",
            "user_name": "Jordan Lee",
            "groupId": "1"
        }
    ],
    "2": [
        {
            "id": "2a",
            "added_by": "admin",
            "joined_at": "2024-02-01T11:00:00Z",
            "user_id": "user_202",
            "user_name": "Casey Brown",
            "groupId": "2"
        },
        {
            "id": "2b",
            "added_by": "user_202",
            "joined_at": "2024-02-03T13:25:00Z",
            "user_id": "user_303",
            "user_name": "Morgan Davis",
            "groupId": "2"
        },
        {
            "id": "2c",
            "added_by": "user_202",
            "joined_at": "2024-02-08T15:40:00Z",
            "user_id": "user_404",
            "user_name": "Riley Miller",
            "groupId": "2"
        }
    ],
    "3": [
        {
            "id": "3a",
            "added_by": "admin",
            "joined_at": "2024-01-10T08:00:00Z",
            "user_id": "user_505",
            "user_name": "Drew Garcia",
            "groupId": "3"
        },
        {
            "id": "3b",
            "added_by": "user_505",
            "joined_at": "2024-01-12T12:30:00Z",
            "user_id": "user_606",
            "user_name": "Patel Williams",
            "groupId": "3"
        },
        {
            "id": "3c",
            "added_by": "user_505",
            "joined_at": "2024-01-18T17:10:00Z",
            "user_id": "user_707",
            "user_name": "Cameron Jones",
            "groupId": "3"
        },
        {
            "id": "3d",
            "added_by": "user_606",
            "joined_at": "2024-01-25T14:00:00Z",
            "user_id": "user_808",
            "user_name": "Blake Martinez",
            "groupId": "3"
        },
        {
            "id": "3e",
            "added_by": "user_707",
            "joined_at": "2024-02-02T10:45:00Z",
            "user_id": "user_909",
            "user_name": "Jordan Taylor",
            "groupId": "3"
        }
    ]
}

# id: id,
# added_by: added_by,
# joined_at: joined_at,
# user_id: user_id,
# user_name: user_name,
# groupId: groupId
@app.route('/api/group/members', methods=['GET'])
def get_members():
    group = request.args.get("group-id")
    m = members[group]
    return jsonify(m)

if __name__ == "__main__":
    app.run(debug=True)
