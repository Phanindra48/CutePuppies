var express = require('express');
var stormpath = require('express-stormpath');

var app = express();

app.set('views', './views');
app.set('view engine', 'jade');

var stormpathMiddleware = stormpath.init(app, {
  apiKeyFile: 'C:\\Users\\ppydisetty\\Documents\\GitHub\\CutePuppies\\apiKey.properties',
  application: 'https://api.stormpath.com/v1/applications/4pYa9wtcIPe7PCmKPh0t1A',
  secretKey: 'GYrHnNdIwrYIPRalAFX1GdXMFAHpGcOjJS5U3VA',
  expandCustomData: true,
  enableForgotPassword: true
});

app.use(stormpathMiddleware);

app.get('/', function(req, res) {
  res.render('home', {
    title: 'Welcome'
  });
});

app.listen(3000);