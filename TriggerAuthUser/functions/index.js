const functions = require("firebase-functions");
const admin = require("firebase-admin");

admin.initializeApp();

// Handles creates for a user on sdk side
exports.createUserProfile = functions.auth.user().onCreate(async (user) => {
    const db = admin.firestore();

    const userDoc = {
        uid: user.uid,
        email: user.email || null,
        created_at: new Date(),
        role: "user"
    };

    try {
        await db.collection("users").doc(user.uid).set(userDoc);
        console.log(`User profile created for UID ${user.uid} with ${user.email}`);
    } catch (error) {
        console.error("Error writing user to Firestore:", error);
    }
});

// Handles deletes for a user on SDK side
exports.deleteUserProfile = functions.auth.user().onDelete(async (user) => {
    const db = admin.firestore();

    try {
        await db.collection("users").doc(user.uid).delete();
        console.log(`User profile deleted for UID ${user.uid} with email ${user.email}`);
    } catch (error) {
        console.error("Error deleting user from Firestore:", error);
    }
});
