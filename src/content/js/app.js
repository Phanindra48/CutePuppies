
$(document).ready(function(){
  cutePuppies.init();
})

var cutePuppies = cutePuppies || {};
cutePuppies = (function(){
  var globals = {
    upvotes:[],
    downvotes:[],
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
      alert('dsfsdf');
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