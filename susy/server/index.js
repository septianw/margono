var express = require('express');
var app = express();
var exec = require('child_process').exec;

app.get('/:'+/[a-z]/, function (req, res) {
  var user = req.params.user
  console.log(req.params);
  if (user) {
    exec('/home/asep/gocode/src/github.com/septianw/margono/susy/susy r '+/[a-z]/, function (err, stdout, stderr) {
      res.send(stdout);
    });
  } else {
    res.send('no');
  }
});

app.listen(4545, function() {
  console.log('App listen on port 4545');
});
