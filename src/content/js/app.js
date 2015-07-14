
$(document).ready(function(){
  cutePuppies.init();
})

var cutePuppies = cutePuppies || {};
cutePuppies = (function(){
  var globals = {
    upvotes:[],
    downvotes:[],
    currentPage:"1",
    user_id:0
  }
  var functions = {
    init:function(){
      $('.allPuppies').off().on('click',function(){
        functions.getAllPuppies();
      });
      $('.topPuppies').off().on('click',function(){
        functions.getTopPuppies();
      });
      globals.user_id = $('#user_id').val();
      $('.appLogout').off().on('click',function(){
        $.post('/logout',function(data){
          window.location.href = '/'
        });
      });
      functions.getAllPuppies();
    },
    bindButtons:function(){
      /*
      $('.page-content').off().on('click','.puppyLike',function(e){
        functions.like();
      });

      $('.page-content').off().on('click','a.puppyDisLike',function(e){
        functions.dislike();
      });

      $('.page-content').off().on('click','a.puppy',function(e){
        functions.enlargeImage();
      });
      */
      $('.puppyLike').off().on('click',function(e){
        functions.like();
      });

      $('.puppyDisLike').off().on('click',function(e){
        functions.dislike();
      });

      $('.puppy').off().on('click',function(e){
        functions.enlargeImage();
      });

    },
    getAllPuppies:function(){
      var params = {
          "page" : '1',
          "uid": parseInt(globals.user_id)
        };
      var jsonStr = JSON.stringify(params)
      console.log(jsonStr);
      $.ajax({
        contentType: "application/json; charset=utf-8",
        type: 'post',
        url: '/pups',
        data:jsonStr,
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
          functions.bindButtons();
        }
        })
        .fail(function (jqXHR, textStatus, errorThrown) {
            alert('getAllPuppies -> ' + errorThrown);
        });
    },
    getTopPuppies:function(){
      var params = {
          "page" : '1',
          "uid": parseInt(globals.user_id)
        };
      var jsonStr = JSON.stringify(params)
      $.ajax({
        contentType: "application/json; charset=utf-8",
        type: 'post',
        url: '/top',
        data:jsonStr,
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
          $('.page-content').html(data);
          $('html, body').animate({scrollTop: '0px'}, 300);
          functions.bindButtons();
        }
        })
        .fail(function (jqXHR, textStatus, errorThrown) {
            alert('Top puppies -> ' + errorThrown);
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

        puppyFactory.votePuppy(puppyId,1,globals.user_id);
      }
    },
    dislike:function() {
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
        puppyFactory.votePuppy(puppyId,0,globals.user_id);
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
  var puppyUrlBase = '/updatevotes';
  var factory = {
    votePuppy: function(id,choice,uid){
      console.log(id,choice,uid);
      var vote = {};
      vote.ID = id;
      vote.VT = choice;
      vote.UID = parseInt(uid);
      var data = JSON.stringify(vote);
      console.log(data);
      $.ajax({
        url: puppyUrlBase,
        contentType: "application/json; charset=utf-8",
        data:data,
        type: 'post',
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