/**
 * @author Gennaro Cuomo
 * @author Ethan Coeytaux
 * 
 * Handles routing for all website pages for Chain Vote including: Login Page, Topics Page, and Voting Page.
 */

var express = require('express');
var router = express.Router();
var url = require('url');
var session = require('express-session');
var chaincode = require('../libs/blockchainSDK');

// Loads login page.
router.get('/', function (req, res, next) {
  res.render('login', { title: 'Chain Vote' });
});

// Submits username and routes user to main topic page.
router.get('/topics', function (req, res) {
  // If user doesnt not have a session id redirect them to the login page.
  if(!req.session.name || req.session.name == null) {
    res.redirect('/');
  }
  res.render('topic-select', { title: 'Chain Vote' });
});

// Routes user to selected topic page.
router.get('/topic/:id', function (req, res) {
  // If user doesnt not have a session id redirect them to the login page.
  if(!req.session.name || req.session.name == null) {
    res.redirect('/');
  }
  // Use the string query to aquire the topic id.
  var url_parts = url.parse(req.url, true);
  var id;
  for (var i in url_parts.query) {
    id = url_parts.query[i];
  }
  var args = [];
  args.push(id);
  args.push(req.session.name);
  // Querry tyhe chaincode for the topic id.
  chaincode.query('get_topic', args, function (err, data) {
    if (data && !err) {
      //Send successful response with topic data.
      res.render('topic', { title: 'Chain Vote', topicName: data.Topic.topic, topicID: data.Topic.topic_id });
    } else {
      // Send bad response with err message.
      res.render('topic', { title: 'Chain Vote', topicName: "TOPIC NOT FOUND", topicID: "" }); //TODO return topic not found object?
    }
  });
});

// Routing for manager page.
router.get('/manager', function (req, res) {
  //TODO add check to see if user is manager.
  // This might do that but unsure.
  //if(req.session.name.includes('manager')){
  res.render('manager', { title: 'Chain Vote' });
  //} else {
  //  res.redirect('/');
  //}
});

// Logs out the user and wipes their session data.
router.get('/logout', function (req, res) {
  req.session.name = null;
  res.redirect('/');
});

module.exports = router;
