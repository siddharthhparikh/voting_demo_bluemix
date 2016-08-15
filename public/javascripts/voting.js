/**
 * @author Gennaro Cuomo
 * @author Ethan Coeytaux
 * 
 * Handles voting events including the remaining vote count.
 */

$(document).ready(function () {
  var maxVotes;
  $.get('/api/get-account', function (data, status) {
    if (data) {
      maxVotes = data.vote_count;
      $('#remaining-votes').append(maxVotes);
      $('.hidden').hide();
    }
    return 0;
  });

  //
  // Get current topic info
  //

  // Query the server for a the topic so that it can be loaded to the page
  $.get('/api/get-topic', { 'topicID': $('#topicID').html() }, function (data, status) {
    // If there is a response.
    if (data) {
      if (data.Status == "open") {
        // Create candidates
        data.Topic['choices[]'].forEach(function (entry) {
          $('#candidates tr:last').after('<tr><td>' + entry + '</td><td><input type="number" class="votes" min="0"/></td></tr>')
        });
        $('.votes').val('0');
      } else if (data.Status == "closed" || data.Status == "voted") {
        $('.vote-header').hide();

        var graphData = [];

        for (var i = 0; i < data.Topic['choices[]'].length; i++) {
          graphData.push([data.Topic['choices[]'][i], parseInt(data.Topic['votes[]'][i])]);
        }

        $('#content-block').highcharts({
          chart: {
            plotBackgroundColor: null,
            plotBorderWidth: 0,
            plotShadow: false
          },
          title: {
            text: 'RESULTS',
            align: 'center',
            verticalAlign: 'middle',
            y: 50
          },
          tooltip: {
            pointFormat: '{series.name}: <b>{point.percentage:.1f}%</b>'
          },
          plotOptions: {
            pie: {
              dataLabels: {
                enabled: true,
                distance: -50,
                style: {
                  fontWeight: 'bold',
                  color: 'white',
                  textShadow: '0px 1px 2px black'
                }
              },
              startAngle: -90,
              endAngle: 90,
              center: ['50%', '75%']
            }
          },
          series: [{
            type: 'pie',
            name: data.Topic.topic,
            innerSize: '50%',
            data: graphData
          }]
        });
      }
    }
  });

  var currentVotesCast = 0; //represents sum of votes in voting boxes

  //
  // Submit user votes
  //
  $('#submit').click(function (e) {
    e.preventDefault(e);

    if (currentVotesCast > maxVotes) {
      alert('You cannot cast more votes that you have available to you, please fix this and try again.');
    } else {
      $.get('/api/get-topic', { "topicID": $('#topicID').html() }, function (data, status) {
        if (data) {
          data = data.Topic;

          var votesArray = [];
          var votes = document.getElementsByClassName('votes');
          $('.votes').each(function () {
            var val = this.value.toString();
            if (val == "") val = "0";
            console.log("VAL: ", val);
            votesArray.push(val);
          });

          var voteJSON = {
            "topic": data.topic_id,
            "choices[]": data["choices[]"],
            "votes[]": votesArray,
            "voter": null, //TODO this should be username
            "castDate": (new Date()).toString() //TODO should this be done on chaincode side of things?
          }

          // Submit the vote object to the server.
          $.post('/api/vote-submit', voteJSON, function (data, status) {
            // Handle response
            data = JSON.parse(data);
            if (data.status == 'success') {
              console.log('Votes Submitted');
              // Reroute the user back to the home page.
              window.location.replace('/topics');
            } else {
              console.log('Error: ' + data.status);
            }
          });
        }
      });
    }
  });

  // Remaining votes
  $(document).on('change', '.votes', function (e) {
    e.preventDefault();
    var sum = 0;
    // Collect sum of all votes applied.
    $('.votes').each(function () {
      var index = $(".votes").index(this);
      sum += parseInt($(this).val());
    });

    currentVotesCast = sum;

    $('#remaining-votes').html(maxVotes - sum);
    if (maxVotes >= sum) {
      document.getElementById("remaining-votes").style.color = "#005C34";
      //document.getElementById("remaining-votes-header").style.color = "#005C34";
    } else {
      document.getElementById("remaining-votes").style.color = "#C20900";
      //document.getElementById("remaining-votes-header").style.color = "#C20900";
    }
  });

  $('#title').click(function () {
    window.location.replace('../topics');
  });
});



