
$(document).ready(function () {
  $.get('/api/manager', function (data, status) {
    if (data) {
      // Create a <tr> for each request that exists.
      data.AllAccReq.forEach(function (entry) {
        // Generate and append the new request html.        
        console.log(entry.privileges[0])
        // line of death
        $('#request-table tr:last').after(
          '<tr class="request"><td>' + entry.name +
          '</td><td>' + entry.email +
          '</td><td>' + entry.org +
          '</td></td><td>' + entry.req_time +
          '</td><td>' +
          //'<select class="privilege" name="priv" value="creator">' +
          //'<?php $status = ;?>' +
          '<select required class="privilege">' +
          //'<select required class="privilege" data-bind="kendoDropDownList: { data: menuOptions, dataTextField: "option", dataValueField: "menu", value: entry.privileges[0]} />' +
          '<option value="default">Default</option>' +
          '<option value="creator">Creator</option>' +
          '<option value="manager">Manager</option>' +
          '</select>' +
          '</td><td><input type"number" min="0" class="vote-amount request-info" value="5"/></td><td>' +
          '<i class="button approve material-icons" name="' + entry.name +
          '" email="' + entry.email +
          '" org="' + entry.org +
          '" priv="">check</i><i class="button decline material-icons" name="' + entry.name +
          '" email="' + entry.email +
          '" org="' + entry.org + '"priv"">close</i></td></tr>');
      });

    }
  });
  // Events for the approve/decline buttons.
  $(document).on('click', '.approve', function () {
    console.log($(this).parent().parent().find('.vote-amount').val());
    $(this).parent().parent().fadeOut();
    var user = {
      Name: $(this).attr("name"),
      VoteCount: $(this).parent().parent().find('.vote-amount').val(),
      Email: $(this).attr("email"),
      Org: $(this).attr("org"),
      Privileges: $(this).parent().parent().find(":selected").text().toLowerCase()
    }
    console.log("inside on click approve");
    $.post('/api/approved', user, function (data, status) {
      //$(this).parent().remove();
      console.log("after finishing approving");
      location.reload();
    });
  });

  $(document).on('click', '.decline', function () {
    $(this).parent().parent().fadeOut();
    //TODO Maybe send notification to user.
    var user = {
      Name: $(this).attr("name"),
      VoteCount: $(this).parent().parent().find('.vote-amount').val(),
      Email: $(this).attr("email"),
      Org: $(this).attr("org"),
      Privileges: $(this).parent().parent().find(":selected").text().toLowerCase()
    }
    console.log("inside on click decline");
    $.post('/api/declined', user, function (data, status) {
      //$(this).parent().remove();
      console.log("after finishing declining");
      location.reload();
    });
  });
  $('#title').click(function () {
    window.location.replace('../topics');
  });
});
