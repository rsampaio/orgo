var account_id;

// http://stackoverflow.com/a/18652401
function getCookie(key) {
  var keyValue = document.cookie.match('(^|;) ?' + key + '=([^;]*)(;|$)');
  return keyValue ? keyValue[2] : null;
}

function deleteCookie( name ) {
  document.cookie = name + '=; expires=Thu, 01 Jan 1970 00:00:01 GMT;';
}

function onSignIn(authRequest) {
  $.post("/google/oauth", {code: authRequest.code},
         function(data, statusText, request) {
           console.log(statusText);
           if (statusText == 'success') {
             window.location = '/';
           }
         });
}

function signOut() {
  var auth2 = gapi.auth2.getAuthInstance();
  auth2.signOut().then(function () {
    console.log('User signed out.');
  });
  deleteCookie("orgo-session");
  window.location = '/';
}

function startApp() {
  console.log("start app");
  gapi.load('auth2', function() {
    auth2 = gapi.auth2.init({
      client_id: "203571506393-0vg0i4muh04j9t3vurm68c867ht9uccl.apps.googleusercontent.com",
      scope: "profile email https://www.googleapis.com/auth/calendar"
    });

    $('#google-logout')
      .text("Sign-out")
      .click(signOut);

    $('#google-login').click(function() {
      auth2.grantOfflineAccess({redirect_uri: "postmessage"}).then(onSignIn);
    });

    // Sign the user in, and then retrieve their ID.
    /*
    auth2.isSignedIn.listen(function(signedIn) {
      if (signedIn) {
        account_id = auth2.currentUser.get().getId();
        access_token = auth2.currentUser.get().getAuthResponse().access_token;
        id_token = auth2.currentUser.get().getAuthResponse().id_token;

        $('#google-logout').text("Sign-out");

        session = getCookie("orgo-session");
        if (session == null) {
          $.post("/google/verify_token",
                 {
                   account_id: account_id,
                   access_token: access_token,
                   id_token: id_token
                 },
                 function(data, url, request) {
                   console.log(request);
                 });
        }
      } else {
        console.log("not authenticated");
        $('#google-logout').text("");
      }
    });
    */
  });
}
