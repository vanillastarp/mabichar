//delchar 刪除角色
function delchar(action) {
    $('#ModalTitle').html('刪除角色');
    $('#ModalContent').html('確定要刪除此角色?');
    if (action == true) {
        var form = $('[name="delcharForm"]');
        form.submit();
    }
}