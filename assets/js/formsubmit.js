/**
A generic form submission script that turns formdata into JSON and posts it.

Copyright 2016 the connectordb team.
Licensed under the MIT license
**/

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

function submitForm(formid) {
	  event.preventDefault();
	  console.log("submitting");
	var form = $(formid) //document.getElementById(formid)
	var formdata = form.serializeObject()
	console.log(formdata)
    $.ajax({
       type: "POST",
       url: form.attr('action'),
       data: JSON.stringify(formdata), // serializes the form's elements.
       success: function(data)
       {
		   window.reload();
       }
   }).done(function(){
	   window.reload();
   }).fail(function(){
	   alert("Update failed");
   }).always(function(){
	   console.log("form submitted");
   });
	return false;
}
