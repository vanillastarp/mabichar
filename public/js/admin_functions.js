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