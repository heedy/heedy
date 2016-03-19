/**
A generic form submission script that turns formdata into JSON and posts it.

Copyright 2016 the connectordb team.
Licensed under the MIT license
**/

if (!String.prototype.endsWith) {
  String.prototype.endsWith = function(searchString, position) {
      var subjectString = this.toString();
      if (typeof position !== 'number' || !isFinite(position) || Math.floor(position) !== position || position > subjectString.length) {
        position = subjectString.length;
      }
      position -= searchString.length;
      var lastIndex = subjectString.indexOf(searchString, position);
      return lastIndex !== -1 && lastIndex === position;
  };
}


$.fn.serializeObject = function()
{
    var o = {};
    var a = this.serializeArray();
    $.each(a, function() {
        if (o[this.name] !== undefined) {
            if (!o[this.name].push) {
                o[this.name] = [o[this.name]];
            }
            o[this.name].push(this.value || '');
        } else {
            o[this.name] = this.value || '';
        }
    });
    return o;
};

function submitForm(formid, actionAcceptField) {
	var form = $(formid); //document.getElementById(formid)
	var formdata = form.serializeObject()

    var action = form.attr('action');
    if(actionAcceptField) {
        if(! action.endsWith("/"))
            action += "/";
        action += formdata[actionAcceptField];
    }

    console.log(action);

    return submissionHelper(formid, "POST", action);
}

function submitUpdate(formid) {
    return submissionHelper(formid, "PUT");
}

function submissionHelper(formid, method, action) {
	  event.preventDefault();
	  console.log("submitting");
	var form = $(formid) //document.getElementById(formid)
	var formdata = form.serializeObject()
	console.log(formdata)
    console.log(method);

    var fullaction = action || form.attr('action');
    $.ajax({
       type: method,
       url: fullaction,
       data: JSON.stringify(formdata), // serializes the form's elements.
       success: function(data)
       {
		   window.location.reload(true);
       }
   }).done(function(){
	   window.location.reload();
   }).fail(function(){
	   alert("Update failed");
   }).always(function(){
	   console.log("form submitted");
   });
	return false;
}
