package users

import(
    "fmt"
    "net/http"
    )


var firstrunpage = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta http-equiv="X-UA-Compatible" content="IE=edge">
<link rel="stylesheet" href="https://maxcdn.bootstrapcdn.com/bootstrap/3.3.2/css/bootstrap.min.css">
<link href="https://maxcdn.bootstrapcdn.com/bootstrap/3.3.2/css/bootstrap-theme.min.css" rel="stylesheet">
<script src="//code.jquery.com/jquery-1.11.2.min.js"></script>
<title>Firstrun</title>
<script type="text/javascript">

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
                    grantAdmin(data.Unsanitized[0]);
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
            }
            },
            dataType: 'jsonp'
            });
}

</script>
</head>

<body>
<div class="container">
<h1>Welcome</h1>

<p>In order to get going, you'll need to set up a username and password for the
first user in the system, afterwards you can activate all the builtin modules.</p>

<form class="form-signin" onsubmit="return checkForm(this);">
<h2 class="form-signin-heading">Create a user</h2>
<label for="inputEmail" class="sr-only">Email address</label>
<input type="email" id="inputEmail" class="form-control" placeholder="Email address" name="email" required autofocus>
<label for="inputPassword" class="sr-only">Username</label>
<input type="text" id="inputUsername" class="form-control" placeholder="Username" name="username" required>

<label for="inputPassword" class="sr-only">Password</label>
<input type="password" id="inputPassword" class="form-control" placeholder="Password" name="password1" required>

<label for="inputPassword" class="sr-only">Password (Repeat)</label>
<input type="password" id="inputPassword2" class="form-control" placeholder="Password" name="password2" required>
<button class="btn btn-lg btn-primary btn-block" type="submit">Create</button>
</form>

</div> <!-- /container -->
</body>
</html>`


func firstRunHandler(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, firstrunpage)
}
