/*
    Copyright (c) 2016 VMware, Inc. All Rights Reserved.
    Licensed under the Apache License, Version 2.0 (the "License");
    you may not use this file except in compliance with the License.
    You may obtain a copy of the License at
        
        http://www.apache.org/licenses/LICENSE-2.0
        
    Unless required by applicable law or agreed to in writing, software
    distributed under the License is distributed on an "AS IS" BASIS,
    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
    See the License for the specific language governing permissions and
    limitations under the License.
*/
var global_messages = {
  "username_is_required" : {
	 "en-US": "Username is required.",
	 "zh-CN": "用户名为必填项。"
  },
  "username_has_been_taken" : {
     "en-US": "Username has been taken.",
	 "zh-CN": "用户名已被占用。"
  },
  "username_is_too_long" : {
	 "en-US": "Username is too long. (maximum 20 characters)",
	 "zh-CN": "用户名长度超出限制。（最长为20个字符）"
  },
  "username_contains_illegal_chars": {
	 "en-US": "Username contains illegal character(s).",
	 "zh-CN": "用户名包含不合法的字符。"
  },
  "email_is_required" : {
	 "en-US": "Email is required.",
	 "zh-CN": "邮箱为必填项。"
  },
  "email_contains_illegal_chars" : {
	 "en-US": "Email contains illegal character(s).",
	 "zh-CN": "邮箱包含不合法的字符。"
  },
  "email_has_been_taken" : {
	 "en-US": "Email has been taken.",
	 "zh-CN": "邮箱已被占用。"
  },
  "email_content_illegal" : {
	 "en-US": "Email format is illegal.",
	 "zh-CN": "邮箱格式不合法。"
  },
  "email_does_not_exist" : {
	 "en-US": "Email does not exist.",
	 "zh-CN": "邮箱不存在。"
  },
  "realname_is_required" : {
	 "en-US": "Full name is required.",
	 "zh-CN": "全名为必填项。"
  },
  "realname_is_too_long" : {
	 "en-US": "Full name is too long. (maximum 20 characters)",
	 "zh-CN": "全名长度超出限制。（最长为20个字符）"
  },
  "realname_contains_illegal_chars" : {
	 "en-US": "Full name contains illegal character(s).",
	 "zh-CN": "全名包含不合法的字符。"
  },
  "password_is_required" : {
	 "en-US": "Password is required.",
	 "zh-CN": "密码为必填项。"
  },
  "password_is_invalid" : {
	 "en-US": "Password is invalid. At least 7 characters with 1 lowercase letter, 1 capital letter and 1 numeric character.",
	 "zh-CN": "密码无效。至少输入 7个字符且包含 1个小写字母，1个大写字母和 1个数字。"
  },
  "password_is_too_long" : {
	 "en-US": "Password is too long. (maximum 20 characters)",
	 "zh-CN": "密码长度超出限制。（最长为20个字符）"
  },
  "password_does_not_match" : {
	 "en-US": "Passwords do not match.",
	 "zh-CN": "两次密码输入不一致。"
  },
  "comment_is_too_long" : {
	 "en-US": "Comment is too long. (maximum 20 characters)",
	 "zh-CN": "备注长度超出限制。（最长为20个字符）"
  },
  "comment_contains_illegal_chars" : {
	 "en-US":  "Comment contains illegal character(s).",
	 "zh-CN": "备注包含不合法的字符。"
  },
  "project_name_is_required" : {
	 "en-US": "Project name is required.",
	 "zh-CN": "项目名称为必填项。"
  },
  "project_name_is_too_short" : {
	 "en-US": "Project name is too short. (minimum 4 characters)",
	 "zh-CN": "项目名称至少要求 4个字符。"
  },
  "project_name_is_too_long" : {
	 "en-US": "Project name is too long. (maximum 30 characters)",
	 "zh-CN": "项目名称长度超出限制。（最长为30个字符）"
  },
  "project_name_contains_illegal_chars" : {
	 "en-US": "Project name contains illegal character(s).",
	 "zh-CN": "项目名称包含不合法的字符。"
  },
  "project_exists" : {
	 "en-US": "Project exists.",
	 "zh-CN": "项目已存在。"
  },
  "delete_user" : {
	 "en-US": "Delete User",
	 "zh-CN": "删除用户"
  },	
  "are_you_sure_to_delete_user" : {
	 "en-US": "Are you sure to delete ",
	 "zh-CN": "确认要删除用户 "
  },
  "input_your_username_and_password" : {
	 "en-US": "Please input your username and password.",
	 "zh-CN": "请输入用户名和密码。"
  },
  "check_your_username_or_password" : {
	 "en-US": "Please check your username or password.",
	 "zh-CN": "请输入正确的用户名或密码。"
  },
  "title_login_failed" : {
	 "en-US": "Login Failed",
	 "zh-CN": "登录失败"
  },
  "title_change_password" : {
	 "en-US": "Change Password",
	 "zh-CN": "修改密码"
  },
  "change_password_successfully" : {
	 "en-US": "Password changed successfully.",
	 "zh-CN": "密码已修改。"
  },
  "title_forgot_password" : {
     "en-US": "Forgot Password",
	 "zh-CN": "忘记密码"
  },
  "email_has_been_sent" : {
	 "en-US": "Email for resetting password has been sent.",
	 "zh-CN": "重置密码邮件已发送。"
  },
  "send_email_failed" : {
	 "en-US": "Failed to send Email for resetting password.",
	 "zh-CN": "重置密码邮件发送失败。"
  },
  "please_login_first" : {
	 "en-US": "Please login first.",
	 "zh-CN": "请先登录。"
  },
  "old_password_is_not_correct" : {
	 "en-US": "Old password is not correct.",
	 "zh-CN": "原密码输入不正确。"
  },
  "please_input_new_password" : {
	 "en-US": "Please input new password.",
	 "zh-CN": "请输入新密码。"
  },
  "invalid_reset_url": {
	 "en-US": "Invalid URL for resetting password.",
	 "zh-CN": "无效密码重置链接。"
  },
  "reset_password_successfully" : {
	 "en-US": "Reset password successfully.",
	 "zh-CN": "密码重置成功。"
  },
  "internal_error": {
	 "en-US": "Internal error.",
	 "zh-CN": "内部错误，请联系系统管理员。"
  },
  "title_reset_password" : {
	 "en-US": "Reset Password",
	 "zh-CN": "重置密码"
  },
  "title_sign_up" : {
	 "en-US": "Sign Up",
	 "zh-CN": "注册"
  },
  "title_add_user": {
     "en-US": "Add User",
     "zh-CN": "新增用户"  
  },
  "registered_successfully": {
	 "en-US": "Signed up successfully.",
	 "zh-CN": "注册成功。"
  },
  "registered_failed" : {
	 "en-US": "Failed to sign up.",
	 "zh-CN": "注册失败。"
  },
  "added_user_successfully": {
     "en-US": "Added user successfully.",
     "zh-CN": "新增用户成功。"  
  },
  "added_user_failed": {
     "en-US": "Added user failed.",
     "zh-CN": "新增用户失败。"  
  },
  "projects" :  {
	 "en-US": "Projects",
	 "zh-CN": "项目"
  },
  "repositories" : {
	 "en-US": "Repositories",
	 "zh-CN": "镜像仓库"
  },
  "no_repo_exists"  :{
     "en-US": "No repositories found, please use 'docker push' to upload images.",
	 "zh-CN": "未发现镜像，请用‘docker push’命令上传镜像。"
  },
  "tag" : {
     "en-US": "Tag",
	 "zh-CN": "标签"
  },
  "pull_command": {
	 "en-US": "Pull Command",
	 "zh-CN": "Pull 命令"
  },
  "image_details" : {
	 "en-US": "Image Details",
	 "zh-CN": "镜像详细信息"
  },
  "add_members" : {
	 "en-US": "Add Member",
	 "zh-CN": "添加成员"
  },
  "edit_members" : {
	 "en-US": "Edit Member",
	 "zh-CN": "编辑成员"
  },
  "add_member_failed" : {
	 "en-US": "Adding Member Failed",
	 "zh-CN": "添加成员失败"
  },
  "please_input_username" : {
	 "en-US": "Please input a username.",
	 "zh-CN": "请输入用户名。"
  },
  "please_assign_a_role_to_user" : {
	 "en-US": "Please assign a role to the user.",
	 "zh-CN": "请为用户分配角色。"
  },
  "user_id_exists" : {
	 "en-US": "User is already a member.",
	 "zh-CN": "用户已经是成员。"
  },
  "user_id_does_not_exist" : {
	 "en-US": "User does not exist.",
	 "zh-CN": "不存在此用户。"
  },
  "insufficient_privileges" : {
	 "en-US": "Insufficient privileges.",
	 "zh-CN": "权限不足。"
  },
  "operation_failed" : {
	 "en-US": "Operation Failed",
	 "zh-CN": "操作失败"
  },
  "button_on" : {
     "en-US": "On",
	 "zh-CN": "打开"	
  },
  "button_off" : {
     "en-US": "Off",
	 "zh-CN": "关闭"	
  }
};