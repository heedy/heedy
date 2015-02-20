
function checkForm(form)
{
    if(form.username.value == "") {
        alert("Error: Username cannot be blank!");
        form.username.focus();
        return false;
    }
    re = /^\w+$/;
    if(!re.test(form.username.value)) {
        alert("Error: Username must contain only letters, numbers and underscores!");
        form.username.focus();
        return false;
    }

    if(form.password1.value == "" || form.password1.value != form.password2.value) {
        alert("Passwords must match and not be blank");
        form.password1.focus();
        return false;
    }

    // do request to create user

    var jsondata = {Id:0,
        Name:form.username.value,
        Email:form.email.value,
        Password:form.password1.value,
        PasswordSalt:"",
        PasswordHashScheme:"",
        Admin:true,
        Phone:"",
        PhoneCarrier:0,
        UploadLimit_Items:999999999,
        ProcessingLimit_S:86400,
        StorageLimit_Gb:1000
    };

    var data = JSON.stringify(jsondata);

    $.ajax({
        type: "POST",
        url: "/api/v1/json/user/",
        data: data,
        success: function(data){console.log(data); created(data, form.username.value);},
        dataType: 'jsonp'
    });


    return false;
}


function created(data, username) {



    $.ajax({
        type: "GET",
        url: "/api/v1/json/" + username + "/",
        data: "",
        success: function(data){
            console.log(data);
            if(data.Status != 200){
                alert("failed");
            } else {
                grantAdmin(data.Users[0]);
            }
        },
        dataType: 'jsonp'
    });

}

function grantAdmin(data) {
    data.Admin = true;
    name = data.Name;
    data = JSON.stringify(data);

    $.ajax({
        type: "PUT",
        url: "/api/v1/json/" + name + "/",
        data: data,
        success: function(data){
            console.log(data);
            if(data.Status != 200){
                alert("admin grant failed");
            } else {
                alert("Admin Granted");
                location.reload();
            }
        },
        dataType: 'jsonp'
    });
}
