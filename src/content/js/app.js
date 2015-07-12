
$(document).ready(function(){
  cutePuppies.init();
})

var cutePuppies = cutePuppies || {};
cutePuppies = (function(){
  var globals = {
    upvotes:[],
    downvotes:[],
    currentPage:"1"
  }
  var functions = {
    init:function(){
      $('.puppyLike').off().on('click',function(){
        functions.like();
      });

      $('.puppyDisLike').off().on('click',function(){
        functions.dislike();
      });

      $('.puppy').off().on('click',function(){
        functions.enlargeImage();
      });

      $('.allPuppies').off().on('click',function(){
        functions.getAllPuppies();
      });
    },
    getAllPuppies:function(){
      $.ajax({
        url: '/pups',
        type: 'GET',
        //dataType: "json",
        data:JSON.stringify({
          page : '1'
        })
      })
      .done(function (data, textStatus, jqXHR) {
        var obj;
        try {
            obj = JSON.parse(data);
        } catch (e) {
            obj = data;
        }
        if (typeof obj.Error != 'undefined' && obj.Error != '') {
          alert(obj.Error);
        }
        else{
          $('.page-content').append(data);
        }
        })
        .fail(function (jqXHR, textStatus, errorThrown) {
            alert('getAllPuppies -> ' + errorThrown);
        });
    },
    like:function(){
      var likeButton = $(event.target);
      var puppyId = likeButton.data('puppy-id');
      var alreadyLiked = globals.upvotes.indexOf(puppyId);
      var alreadyDisLiked = globals.downvotes.indexOf(puppyId);
      if(alreadyLiked * alreadyDisLiked == 1){
        globals.upvotes.push(puppyId);
        var likes = likeButton.find('span#likes').text();
        likeButton.find('span#likes').text(++likes);
        if(!likeButton.hasClass('disabled'))
          likeButton.addClass('disabled');
        var vote = {};
        vote.ID = puppyId;
        vote.VT = true;

        puppyFactory.votePuppy(vote);
      }
    },
    dislike:function(){
      var dislikeButton = $(event.target);
      var puppyId = dislikeButton.data('puppy-id');
      var alreadyLiked = globals.upvotes.indexOf(puppyId);
      var alreadyDisLiked = globals.downvotes.indexOf(puppyId);
      if(alreadyLiked * alreadyDisLiked == 1){
        globals.downvotes.push(puppyId);
        var likes = dislikeButton.find('span#dislikes').text();
        dislikeButton.find('span#dislikes').text(++likes);
        if(!dislikeButton.hasClass('disabled'))
          dislikeButton.addClass('disabled');
        var vote = {};
        vote.ID = puppyId;
        vote.VT = true;

        puppyFactory.votePuppy(vote);
      }
    },
    enlargeImage: function(){
      alert('enlarge image');
    }
  };
  return {
    init : functions.init
  }

})();

var puppyFactory = puppyFactory || {};
puppyFactory = (function(){
  var puppyUrlBase = '/pups';
  var factory = {
    votePuppy: function(vote){
      $.ajax({
        url: puppyUrlBase,
        type: 'PUT',
        success: function(result) {
          // Do something with the result
          console.log('test');
        }
      });
    }
  };

  return {
    votePuppy : factory.votePuppy
  }
})();

/*
function init(){

}
function like(puppy){
  alert('yayy!!');
}
function dislike(puppy){
  alert('oh no!!');
}
function enlargeImage(){

}

*/

/*
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

app.listen(8080);
*/