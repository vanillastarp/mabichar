function delAction(action, path) {
    if($('[name="_method"]').val() == "PUT"){
        $('#ModalTitle').html('刪除');
        $('#ModalContent').html('確定要刪除此筆資料?');
        if(action == true){
            var form = $('[name="EditForm"]');
            var _id = $("#_id").val();
            var actionPath = path + _id;
            $('[name="_method"]').val("DELETE");

            form.attr("action", actionPath);
            form.submit();
        }
    }
}

function GetSkillsToList() {
    $.get("/api/GetSkills", function (data) {
        $(data).each(function (key, val) {
            row =
                "<tr>" +
                "<td scope=\"row\">" + val["skillid"] + "</td>" +
                "<td>" + val["skillName"] + "</td>" +
                "<td>" + val["maxlv"] + "</td>" +
                "<td>" + (val["upgrade"]?"是":"") + "</td>" +
                "<td><a class=\"btn btn-primary\" href=\"/admin/skills/" + val["skillid"] + "/edit\" role=\"button\">編輯</a></td>" +
                "</tr>";
            $("#s" + val["skilltype"] + " tbody").append(row);
        });
    });
}

function GetTalentsToList() {
    $.get("/api/GetTalents", function (data) {
        $(data).each(function (key, val) {
            row =
                "<tr>" +
                    "<td scope=\"row\">" + val["talentid"] + "</td>" +
                    "<td>" + val["category"] + "</td>" +
                    "<td>" + val["talenttitle"] + "</td>" +
                    "<td>" + val["talentlevel"] + "</td>" +
                    "<td><a class=\"btn btn-primary\" href=\"/admin/talentmasters/" + val["talentid"] + "/edit\" role=\"button\">編輯</a></td>" +
                "</tr>";
            $("#s" + val["category"] + " tbody").append(row);
        });
    });
}

function GetTitlesToList() {
    $.get("/api/GetTitles", function (data) {
        $(data).each(function (key, val) {
            row =
                "<tr>" +
                    "<td scope=\"row\">" + val["id"] + "</td>" +
                    "<td>" + val["type"] + "</td>" +
                    "<td>" + val["defaultname"] + "</td>" +
                    "<td>" + val["malename"]+"/"+val["femalename"]+"/"+val["uniname"] + "</td>" +
                    "<td><a class=\"btn btn-primary\" href=\"/admin/titles/" + val["id"] + "/edit\" role=\"button\">編輯</a></td>" +
                "</tr>";
            $("#s" + val["type"] + " tbody").append(row);
        });
    });
}

function GetPetsToList() {
    $.get("/api/GetPets", function (data) {
        $(data).each(function (key, val) {
            row =
                "<tr>" +
                    "<td scope=\"row\">" + val["ID"] + "</td>" +
                    "<td>" + val["type"] + "</td>" +
                    "<td>" + val["localName"] + "</td>" +
                    "<td><a class=\"btn btn-primary\" href=\"/admin/pets/" + val["ID"] + "/edit\" role=\"button\">編輯</a></td>" +
                "</tr>";
            $("#s" + val["type"] + " tbody").append(row);
        });
    });
}

function GetSkillTypesToList() {
    $.get("/api/GetSkillTypes", function (data) {
        $.each(data, function (key, val) {
            skilltype = "<option value=\"" + key + "\">" + val + "</option>";
            $("#inputSkillType").append(skilltype);
        });
        if($("#skilltype").val()!="") $("#inputSkillType").val($("#skilltype").val());
        if($("#race").val()!="") $("#inputRace").val($("#race").val());
        if($("#upgrade").val()!="") $("#inputUpgrade").val($("#upgrade").val());
    });
}

function GetTalentTypesToList() {
    $.get("/api/GetTalentTypes", function (data) {
        $.each(data, function (key, val) {
            talenttype = "<option value=\"" + key + "\">" + val + "</option>";
            $("#inputCategory").append(talenttype);
        });
        if($("#Category").val()!="") $("#inputCategory").val($("#Category").val());
    });
}

function GetTitleTypesToList() {
    $.get("/api/GetTitleTypes", function (data) {
        $.each(data, function (key, val) {
            titletype = "<option value=\"" + key + "\">" + val + "</option>";
            $("#inputType").append(titletype);
        });
        if($("#type").val()!="") $("#inputType").val($("#type").val());
    });
}

function GetPetTypesToList() {
    $.get("/api/GetPetTypes", function (data) {
        $.each(data, function (key, val) {
            titletype = "<option value=\"" + key + "\">" + val + "</option>";
            $("#inputType").append(titletype);
        });
        if($("#type").val()!="") $("#inputType").val($("#type").val());
    });
}