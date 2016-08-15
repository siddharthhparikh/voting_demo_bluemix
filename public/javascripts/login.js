/**
 * @author Gennaro Cuomo
 * 
 * Submits the login request to the server.
 * Routes the user to the topic select page is the login is successful.
 */
$(document).ready(function () {
  $('.hidden').hide();
  console.log('Querying if chaincode has deployed...');
  var intervalVar = setInterval(function() {
  $.get('/api/load-chain', function (data, status) {
    data = JSON.parse(data);
    if (data.status == "success") {
      console.log('Chaincode loaded!');
      clearInterval(intervalVar);
      $('#loading-screen').remove();
      $('#content-header').fadeIn();
      $('#content-block').fadeIn();
      $('#open-register').fadeIn();
    } else {
      console.log('Chaincode failed!');
      $('#loading-screen').fadeIn();
      $('#content-header').hide();
      $('#content-block').hide();
      $('#open-register').hide();
    }
  });
  }, 2000);
  
  //Animation for register info box.
  $('#open-register').click(function() {
    $('#register-box').animate({ height: 'toggle'}, 'fast');
  });
  // Hides menus when user clicks out of them.
  // Hides menus when user clicks out of them.
  $('#master-content').click(function() {
    $('.info-box').fadeOut('fast');
  });

  //
  // Submit user credendials and verify.
  //
  $('#submit').click(function (e) {
    e.preventDefault();

    var user = {
      'account_id': $('#username').val(),
      'password': $('#password').val()
    };
    console.log(user);
    $.post('/api/login', user, function (data, status) {
      data = JSON.parse(data);
      console.log("[DATA]", data);
      // Handle respse "clonse.
      if (data.status === 'success') {
        // Redirect user.
        if(data.type === 'user') {
          window.location.replace("../topics");
        }
        else if(data.type === 'manager') {
          console.log("redirecting to manager...");
          window.location.replace("../manager");
        }
      } else {
        $('#error-msg').html('Error: ' + data.status);
      }
    });
  });


  //
  // Request to register as a new user.
  //
  $('#register-user').click(function() {
    var errFlag = false;
    $('.registration-info').each(function(){
      var index = $(".registration-info").index(this);
      if ($(this).val() == '' && errFlag == false) {
        errFlag = true;
        alert('Error: Input fields can not be left empty.');
      }
    });
    console.log(errFlag)
    if(!errFlag){
      //console.log($('#organization').val());
      // Create request object.
      var newUser = {
        'name': $('#name').val(),
        'email': $('#email').val(),
        'org': $('#organization').val(),
        'privileges':$('#priv-type').val()
      };
      //Send request object.
      console.log(newUser)
      $.post('/api/register', newUser, function (data, status) {
        if (status == 'success') {
          $('#register-box').fadeOut();
          $('#error-msg').html('New account request has been sent.');
        }
      });
    }
  });

  $('#title').click(function() {
    window.location.replace('../topics');
  });
});