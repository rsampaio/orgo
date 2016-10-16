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
    auth2 = gapi.auth2.init({
      client_id: "203571506393-0vg0i4muh04j9t3vurm68c867ht9uccl.apps.googleusercontent.com",
      scope: "profile email https://www.googleapis.com/auth/calendar"
    });

    gapi.signin2.render('google-login', {
        'width': 240,
        'height': 50,
        'longtitle': true,
        'theme': 'dark',
        'onsuccess': onSignIn
    });

    $('#google-logout').click(signOut);

    // Sign the user in, and then retrieve their ID.
    auth2.isSignedIn.listen(function(signedIn) {
      if (signedIn) {
        account_id = auth2.currentUser.get().getId();
        access_token = auth2.currentUser.get().getAuthResponse().access_token;
        id_token = auth2.currentUser.get().getAuthResponse().id_token;
        console.log("account authenticated: " + account_id);
        console.log(access_token)

        $('#google-logout').text("Sign-out");

      } else {
        console.log("not authenticated");
        $('#google-logout').text("");

      }
      $.post("/google/verify_token", {account_id: account_id, access_token: access_token, id_token: id_token}, function(data) {console.log(data)});
    });
  });
}
