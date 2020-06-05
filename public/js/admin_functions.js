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
            //console.log(actionPath)
            //alert(actionPath);
            form.submit();
        }
    }
}

function GetSkillsToList(){
    $.get( "/api/GetSkills", function( data ) {
        $(data).each(function( key,val ) {
            row = 
            "<tr>"+
            "<td scope=\"row\">"+val["skillid"]+"</td>"+
            "<td>"+val["skillName"]+"</td>"+
            "<td>"+val["maxlv"]+"</td>"+
            "<td><a class=\"btn btn-primary\" href=\"/admin/skills/"+val["skillid"]+"/edit\" role=\"button\">編輯</a></td>"+
            "</tr>";
            $("#s"+val["skilltype"]+" tbody").append(row);
        })
    })
}