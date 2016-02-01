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
	 "en-US": "Username is required!",
	 "zh-CN": "用户名为必填项。"
  },
  "username_has_been_taken" : {
     "en-US": "Username has been taken!",
	 "zh-CN": "用户名已被占用！"
  },
  "username_is_too_long" : {
	 "en-US": "Username is too long (maximum is 20 characters)",
	 "zh-CN": "用户名内容长度超出字符数限制。（最长为20个字符）"
  },
  "username_contains_illegal_chars": {
	 "en-US": "Username contains illegal characters.",
	 "zh-CN": "用户名包含不合法的字符。"
  },
  "email_is_required" : {
	 "en-US": "Email is required!",
	 "zh-CN": "邮箱为必填项。"
  },
  "email_contains_illegal_chars" : {
	 "en-US": "Email contains illegal characters.",
	 "zh-CN": "邮箱包含不合法的字符。"
  },
  "email_has_been_taken" : {
	 "en-US": "Email has been taken!",
	 "zh-CN": "邮箱已被占用！"
  },
  "email_content_illegal" : {
	 "en-US": "Email content is illegal.",
	 "zh-CN": "邮箱格式不合法！"
  },
  "email_does_not_exist" : {
	 "en-US": "Email does not exist!",
	 "zh-CN": "邮箱不存在。"
  },
  "realname_is_required" : {
	 "en-US": "Realname is required!",
	 "zh-CN": "全名为必填项。"
  },
  "realname_is_too_long" : {
	 "en-US": "Realname is too long (maximum is 20 characters)",
	 "zh-CN": "全名内容长度超出字符数限制。（最长为20个字符）"
  },
  "realname_contains_illegal_chars" : {
	 "en-US": "Realname contains illegal characters.",
	 "zh-CN": "全名包含不合法的字符。"
  },
  "password_is_required" : {
	 "en-US": "Password is required!",
	 "zh-CN": "密码为必填项。"
  },
  "password_is_invalid" : {
	 "en-US": "Password is invalid. Use more than seven characters with at least one lowercase letter, one capital letter and one numeral.",
	 "zh-CN": "密码无效。至少输入 7个字符且包含 1个小写字母，1个大写字母和数字。"
  },
  "password_is_too_long" : {
	 "en-US": "Password is too long (maximum is 20 characters)",
	 "zh-CN": "密码内容长度超出字符数限制。（最长为20个字符）"
  },
  "password_does_not_match" : {
	 "en-US": "Password does not match the confirmation.",
	 "zh-CN": "两次密码输入内容不一致。"
  },
  "comment_is_too_long" : {
	 "en-US": "Comment is too long (maximum is 20 characters)",
	 "zh-CN": "留言内容长度超过字符数限制。（最长为20个字符）"
  },
  "comment_contains_illegal_chars" : {
	 "en-US":  "Comment contains illegal characters.",
	 "zh-CN": "留言内容包含不合法的字符。"
  },
  "project_name_is_required" : {
	 "en-US": "Project name is required!",
	 "zh-CN": "项目名称为必填项。"
  },
  "project_name_is_too_short" : {
	 "en-US": "Project name is too short (minimum is 4 characters)",
	 "zh-CN": "项目名称内容过于简短。（最少要求4个字符）"
  },
  "project_name_is_too_long" : {
	 "en-US": "Project name is too long (maximum is 30 characters)",
	 "zh-CN": "项目名称内容长度超出字符数限制。（最长为30个字符）"
  },
  "project_name_contains_illegal_chars" : {
	 "en-US": "Project name contains illegal characters.",
	 "zh-CN": "项目名称内容包含不合法的字符。"
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
	 "en-US": "Please check your username or password!",
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
	 "en-US": "Changed password successfully.",
	 "zh-CN": "修改密码操作成功。"
  },
  "title_forgot_password" : {
     "en-US": "Forgot Password",
	 "zh-CN": "忘记密码"
  },
  "email_has_been_sent" : {
	 "en-US": "Email has been sent.",
	 "zh-CN": "重置密码邮件已发送。"
  },
  "send_email_failed" : {
	 "en-US": "Send email failed.",
	 "zh-CN": "邮件发送失败。"
  },
  "please_login_first" : {
	 "en-US": "Please login first!",
	 "zh-CN": "请先登录。"
  },
  "old_password_is_not_correct" : {
	 "en-US": "Old password input is not correct.",
	 "zh-CN": "原密码输入不正确。"
  },
  "please_input_new_password" : {
	 "en-US": "Please input new password.",
	 "zh-CN": "请输入新密码。"
  },
  "invalid_reset_url": {
	 "en-US": "Invalid reset url",
	 "zh-CN": "无效的重置链接"
  },
  "reset_password_successfully" : {
	 "en-US": "Reset password successfully.",
	 "zh-CN": "密码重置成功。"
  },
  "internal_error": {
	 "en-US": "Internal error, please contact sysadmin.",
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
  "registered_successfully": {
	 "en-US": "Registered successfully.",
	 "zh-CN": "注册成功。"
  },
  "registered_failed" : {
	 "en-US": "Registered failed.",
	 "zh-CN": "注册失败。"
  },
  "projects" :  {
	 "en-US": "Projects",
	 "zh-CN": "项目"
  },
  "repositories" : {
	 "en-US": "Repositories",
	 "zh-CN": "镜像资源"
  },
  "no_repo_exists"  :{
     "en-US": "No repositories exist.",
	 "zh-CN": "没有镜像资源。"
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
	 "en-US": "Add Members",
	 "zh-CN": "添加成员"
  },
  "edit_members" : {
	 "en-US": "Edit Members",
	 "zh-CN": "编辑成员"
  },
  "add_member_failed" : {
	 "en-US": "Add Member Failed",
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
	 "en-US": "User ID exists.",
	 "zh-CN": "用户ID已存在。"
  },
  "user_id_does_not_exist" : {
	 "en-US": "User ID does not exist.",
	 "zh-CN": "不存在此用户ID。"
  },
  "insuffient_authority" : {
	 "en-US": "Insufficient authority.",
	 "zh-CN": "权限不足。"
  },
  "operation_failed" : {
	 "en-US": "Operation Failed",
	 "zh-CN": "操作失败"
  },
  "network_error" : {
	 "en-US": "Network Error",
	 "zh-CN": "网络故障"
  },
  "network_error_description" : {
	 "en-US": "Network error, please contact sysadmin.",
	 "zh-CN": "网络故障, 请联系系统管理员。"
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