var account_id;

function onSignIn() {
  var auth2 = gapi.auth2.getAuthInstance();
}

function signOut() {
  var auth2 = gapi.auth2.getAuthInstance();
  auth2.signOut().then(function () {
    console.log('User signed out.');
  });
}

function startApp() {
  console.log("start app");
  gapi.load('auth2', function() {
    auth2 = gapi.auth2.getAuthInstance();

    $('#google-logout').click(signOut);

    // Sign the user in, and then retrieve their ID.
    auth2.isSignedIn.listen(function(signedIn) {
      if (signedIn) {
        account_id = auth2.currentUser.get().getId();
        id_token = auth2.currentUser.get().getAuthResponse().id_token;
        console.log("account authenticated: " + account_id);
        console.log("token_id: " + id_token);

        $('#google-logout').text("Sign-out");

      } else {
        console.log("not authenticated");
        $('#google-logout').text("");

      }
      //$.post("/google/verify_account", {account_id: account_id}, function(data) {console.log(data)});
    });
  });
}
