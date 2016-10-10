var account_id;

function onSignIn() {
  var auth2 = gapi.auth2.getAuthInstance();

  if (!auth2.isSignedIn.get()) {
    auth2.grantOfflineAccess({scope: "profile email https://www.googleapis.com/auth/calendar"}).then(
      function(data) {
        console.log(data);
      }
    );
  } else {
    auth2.signIn();
    //window.location = "/logged.html";
  }
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
      client_id:"203571506393-0vg0i4muh04j9t3vurm68c867ht9uccl.apps.googleusercontent.com",
      scope:"profile email https://www.googleapis.com/auth/calendar"
    });

    $('#google-sign-in').click(onSignIn);
    $('#google-sign-out').click(signOut);

    // Sign the user in, and then retrieve their ID.
    auth2.then(function() {
      if (auth2.isSignedIn.get()) {
        account_id = auth2.currentUser.get().getId();
        console.log("account authenticated: " + account_id);
      } else {
        console.log("not authenticated");
      }
      //$.post("/google/verify_account", {account_id: account_id}, function(data) {console.log(data)});
    });
  });
}
