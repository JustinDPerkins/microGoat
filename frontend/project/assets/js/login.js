function submitLoginForm() {
    var formData = $('#loginForm').serialize();

    // Send AJAX POST request
    $.ajax({
        type: "POST",
        url: "/login",
        data: formData,
        success: function(data, textStatus, xhr) {
            if (xhr.status === 200) {
                // If successful, redirect or do something else
                window.location.href = '/protected';
            } else {
                alert('Login failed. Please try again.');
            }
        },
        error: function(xhr, textStatus, error) {
            alert('Login failed. Please check your credentials and try again.');
        }
    });
}