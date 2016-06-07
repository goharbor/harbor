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
	 "zh-CN": "用户名为必填项。",
	 "ja-JP": "ユーザ名は必須項目です。",
	 "de-DE": "Benutzername erforderlich.",
	 "ru-RU": "Требуется ввести имя пользователя."
  },
  "username_has_been_taken" : {
         "en-US": "Username has been taken.",
	 "zh-CN": "用户名已被占用。",
	 "ja-JP": "ユーザ名はすでに登録されました。",
	 "de-DE": "Benutzername bereits vergeben.",
	 "ru-RU": "Имя пользователя уже используется."
  },
  "username_is_too_long" : {
	 "en-US": "Username is too long. (maximum 20 characters)",
	 "zh-CN": "用户名长度超出限制。（最长为20个字符）",
	 "ja-JP": "ユーザ名が長すぎです。（20文字まで）",
	 "de-DE": "Benutzername ist zu lang. (maximal 20 Zeichen)",
	 "ru-RU": "Имя пользователя слишком длинное. (максимум 20 символов)"
  },
  "username_contains_illegal_chars": {
	 "en-US": "Username contains illegal character(s).",
	 "zh-CN": "用户名包含不合法的字符。",
	 "ja-JP": "ユーザ名に使えない文字が入っています。",
	 "de-DE": "Benutzername enthält ungültige Zeichen.",
	 "ru-RU": "Имя пользователя содержит недопустимые символы."
  },
  "email_is_required" : {
	 "en-US": "Email is required.",
	 "zh-CN": "邮箱为必填项。",
	 "ja-JP": "メールアドレスが必須です。",
	 "de-DE": "E-Mail Adresse erforderlich.",
	 "ru-RU": "Требуется ввести E-mail адрес."
  },
  "email_contains_illegal_chars" : {
	 "en-US": "Email contains illegal character(s).",
	 "zh-CN": "邮箱包含不合法的字符。",
	 "ja-JP": "メールアドレスに使えない文字が入っています。",
	 "de-DE": "E-Mail Adresse enthält ungültige Zeichen.",
	 "ru-RU": "E-mail адрес содержит недопеустимые символы."
  },
  "email_has_been_taken" : {
	 "en-US": "Email has been taken.",
	 "zh-CN": "邮箱已被占用。",
	 "ja-JP": "メールアドレスがすでに使われました。",
	 "de-DE": "E-Mail Adresse wird bereits verwendet.",
	 "ru-RU": "Такой E-mail адрес уже используется."
  },
  "email_content_illegal" : {
	 "en-US": "Email format is illegal.",
	 "zh-CN": "邮箱格式不合法。",
	 "ja-JP": "メールアドレスフォーマットエラー。",
	 "de-DE": "Format der E-Mail Adresse ist ungültig.",
	 "ru-RU": "Недопустимый формат E-mail адреса."
  },
  "email_does_not_exist" : {
	 "en-US": "Email does not exist.",
	 "zh-CN": "邮箱不存在。",
	 "ja-JP": "メールアドレスが存在しません。",
	 "de-DE": "E-Mail Adresse existiert nicht.",
	 "ru-RU": "E-mail адрес не существует."
  },
  "realname_is_required" : {
	 "en-US": "Full name is required.",
	 "zh-CN": "全名为必填项。",
	 "ja-JP": "フルネームが必須です。",
	 "de-DE": "Vollständiger Name erforderlich.",
	 "ru-RU": "Требуется ввести полное имя."
  },
  "realname_is_too_long" : {
	 "en-US": "Full name is too long. (maximum 20 characters)",
	 "zh-CN": "全名长度超出限制。（最长为20个字符）",
	 "ja-JP": "フルネームは長すぎです。（20文字まで）",
	 "de-DE": "Vollständiger Name zu lang. (maximal 20 Zeichen)",
	 "ru-RU": "Полное имя слишком длинное. (максимум 20 символов)"
  },
  "realname_contains_illegal_chars" : {
	 "en-US": "Full name contains illegal character(s).",
	 "zh-CN": "全名包含不合法的字符。",
	 "ja-JP": "フルネームに使えない文字が入っています。",
	 "de-DE": "Vollständiger Name enthält ungültige Zeichen.",
	 "ru-RU": "Полное имя содержит недопустимые символы." 
  },
  "password_is_required" : {
	 "en-US": "Password is required.",
	 "zh-CN": "密码为必填项。",
	 "ja-JP": "パスワードは必須です。",
	 "de-DE": "Passwort erforderlich.",
	 "ru-RU": "Требуется ввести пароль."
  },
  "password_is_invalid" : {
	 "en-US": "Password is invalid. At least 7 characters with 1 lowercase letter, 1 capital letter and 1 numeric character.",
	 "zh-CN": "密码无效。至少输入 7个字符且包含 1个小写字母，1个大写字母和 1个数字。",
	 "ja-JP": "無効なパスワードです。7英数字以上で、 少なくとも１小文字、１大文字と１数字となります。",
	 "de-DE": "Passwort ungültig. Mindestens sieben Zeichen bestehend aus einem Kleinbuchstaben, einem Großbuchstaben und einer Zahl",
	 "ru-RU": "Такой пароль недопустим. Парольл должен содержать Минимум 7 символов, в которых будет присутствовать по меньшей мере 1 буква нижнего регистра, 1 буква верхнего регистра и 1 цифра"
  },
  "password_is_too_long" : {
	 "en-US": "Password is too long. (maximum 20 characters)",
	 "zh-CN": "密码长度超出限制。（最长为20个字符）",
	 "ja-JP": "パスワードは長すぎです。（20文字まで）",
	 "de-DE": "Passwort zu lang. (maximal 20 Zeichen)",
	 "ru-RU": "Пароль слишком длинный (максимум 20 символов)"
  },
  "password_does_not_match" : {
	 "en-US": "Passwords do not match.",
	 "zh-CN": "两次密码输入不一致。",
	 "ja-JP": "確認のパスワードが正しくありません。",
	 "de-DE": "Passwörter stimmen nicht überein.",
	 "ru-RU": "Пароли не совпадают."
  },
  "comment_is_too_long" : {
	 "en-US": "Comment is too long. (maximum 20 characters)",
	 "zh-CN": "备注长度超出限制。（最长为20个字符）",
	 "ja-JP": "コメントは長すぎです。（20文字まで）",
	 "de-DE": "Kommentar zu lang. (maximal 20 Zeichen)",
	 "ru-RU": "Комментарий слишком длинный. (максимум 20 символов)"
  },
  "comment_contains_illegal_chars" : {
	 "en-US":  "Comment contains illegal character(s).",
	 "zh-CN": "备注包含不合法的字符。",
	 "ja-JP": "コメントに使えない文字が入っています。",
	 "de-DE": "Kommentar enthält ungültige Zeichen.",
	 "ru-RU": "Комментарий содержит недопустимые символы."
  },
  "project_name_is_required" : {
	 "en-US": "Project name is required.",
	 "zh-CN": "项目名称为必填项。",
	 "ja-JP": "プロジェクト名は必須です。",
	 "de-DE": "Projektname erforderlich.",
	 "ru-RU": "Необходимо ввести название Проекта."
  },
  "project_name_is_too_short" : {
	 "en-US": "Project name is too short. (minimum 4 characters)",
	 "zh-CN": "项目名称至少要求 4个字符。",
	 "ja-JP": "プロジェクト名は４文字以上です。",
	 "de-DE": "Projektname zu kurz. (mindestens 4 Zeichen)",
	 "ru-RU": "Название проекта слишком короткое. (миниму 4 символа)"
  },
  "project_name_is_too_long" : {
	 "en-US": "Project name is too long. (maximum 30 characters)",
	 "zh-CN": "项目名称长度超出限制。（最长为30个字符）",
	 "ja-JP": "プロジェクト名は長すぎです。（30文字まで）",
	 "de-DE": "Projektname zu lang. (maximal 30 Zeichen)",
	 "ru-RU": "Название проекта слишком длинное (максимум 30 символов)"
  },
  "project_name_contains_illegal_chars" : {
	 "en-US": "Project name contains illegal character(s).",
	 "zh-CN": "项目名称包含不合法的字符。",
	 "ja-JP": "プロジェクト名に使えない文字が入っています。",
	 "de-DE": "Projektname enthält ungültige Zeichen.",
	 "ru-RU": "Название проекта содержит недопустимые символы."
  },
  "project_exists" : {
	 "en-US": "Project exists.",
	 "zh-CN": "项目已存在。",
	 "ja-JP": "プロジェクトはすでに存在しました。",
	 "de-DE": "Projekt existiert bereits.",
	 "ru-RU": "Такой проект уже существует."
  },
  "delete_user" : {
	 "en-US": "Delete User",
	 "zh-CN": "删除用户",
	 "ja-JP": "ユーザを削除",
	 "de-DE": "Benutzer löschen",
	 "ru-RU": "Удалить пользователя"
  },	
  "are_you_sure_to_delete_user" : {
	 "en-US": "Are you sure to delete ",
	 "zh-CN": "确认要删除用户 ",
	 "ja-JP": "ユーザを削除でよろしでしょうか ",
	 "de-DE": "Sind Sie sich sicher, dass Sie folgenden Benutzer löschen möchten: ",
	 "ru-RU": "Вы уверены что хотите удалить пользователя? "
  },
  "input_your_username_and_password" : {
	 "en-US": "Please input your username and password.",
	 "zh-CN": "请输入用户名和密码。",
	 "ja-JP": "ユーザ名とパスワードを入力してください。",
	 "de-DE": "Bitte geben Sie ihr Benutzername und Passwort ein.",
	 "ru-RU": "Введите имя пользователя и пароль."
  },
  "check_your_username_or_password" : {
	 "en-US": "Please check your username or password.",
	 "zh-CN": "请输入正确的用户名或密码。",
	 "ja-JP": "正しいユーザ名とパスワードを入力してください。",
	 "de-DE": "Bitte überprüfen Sie ihren Benutzernamen und Passwort.",
	 "ru-RU": "Проверьте свои имя пользователя и пароль."
  },
  "title_login_failed" : {
	 "en-US": "Login Failed",
	 "zh-CN": "登录失败",
	 "ja-JP": "ログインに失敗しました。",
	 "de-DE": "Anmeldung fehlgeschlagen",
	 "ru-RU": "Ошибка входа"
  },
  "title_change_password" : {
	 "en-US": "Change Password",
	 "zh-CN": "修改密码",
	 "ja-JP": "パスワードを変更します。",
	 "de-DE": "Passwort ändern",
	 "ru-RU": "Сменить пароль"
  },
  "change_password_successfully" : {
	 "en-US": "Password changed successfully.",
	 "zh-CN": "密码已修改。",
	 "ja-JP": "パスワードを変更しました。",
	 "de-DE": "Passwort erfolgreich geändert.",
	 "ru-RU": "Пароль успешно изменен."
  },
  "title_forgot_password" : {
         "en-US": "Forgot Password",
	 "zh-CN": "忘记密码",
	 "ja-JP": "パスワードをリセットします。",
	 "de-DE": "Passwort vergessen",
	 "ru-RU": "Забыли пароль?"
  },
  "email_has_been_sent" : {
	 "en-US": "Email for resetting password has been sent.",
	 "zh-CN": "重置密码邮件已发送。",
	 "ja-JP": "パスワードをリセットするメールを送信しました。",
	 "de-DE": "Eine E-Mail mit einem Wiederherstellungslink wurde an Sie gesendet.",
	 "ru-RU": "На ваш E-mail было выслано письмо с инструкциями по сбросу пароля."
  },
  "send_email_failed" : {
	 "en-US": "Failed to send Email for resetting password.",
	 "zh-CN": "重置密码邮件发送失败。",
	 "ja-JP": "パスワードをリセットするメールを送信する際エラーが出ました",
	 "de-DE": "Fehler beim Senden der Wiederherstellungs-E-Mail.",
	 "ru-RU": "Ошибка отправки сообщения."
  },
  "please_login_first" : {
	 "en-US": "Please login first.",
	 "zh-CN": "请先登录。",
	 "ja-JP": "この先にログインが必要です。",
	 "de-DE": "Bitte melden Sie sich zuerst an.",
	 "ru-RU": "Сначала выполните вход в систему."
  },
  "old_password_is_not_correct" : {
	 "en-US": "Old password is not correct.",
	 "zh-CN": "原密码输入不正确。",
	 "ja-JP": "現在のパスワードが正しく入力されていません。",
	 "de-DE": "Altes Passwort ist nicht korrekt.",
	 "ru-RU": "Старый пароль введен неверно."
  },
  "please_input_new_password" : {
	 "en-US": "Please input new password.",
	 "zh-CN": "请输入新密码。",
	 "ja-JP": "あたらしいパスワードを入力してください",
	 "de-DE": "Bitte geben Sie ihr neues Passwort ein.",
	 "ru-RU": "Пожалуйста, введите новый пароль."
  },
  "invalid_reset_url": {
	 "en-US": "Invalid URL for resetting password.",
	 "zh-CN": "无效密码重置链接。",
	 "ja-JP": "無効なパスワードをリセットするリンク。",
	 "de-DE": "Ungültige URL zum Passwort wiederherstellen.",
	 "ru-RU": "Неверный URL для сброса пароля."
  },
  "reset_password_successfully" : {
	 "en-US": "Reset password successfully.",
	 "zh-CN": "密码重置成功。",
	 "ja-JP": "パスワードをリセットしました。",
	 "de-DE": "Passwort erfolgreich wiederhergestellt.",
	 "ru-RU": "Пароль успешно сброшен."
  },
  "internal_error": {
	 "en-US": "Internal error.",
	 "zh-CN": "内部错误，请联系系统管理员。",
	 "ja-JP": "エラーが出ました、管理者に連絡してください。",
	 "de-DE": "Interner Fehler.",
	 "ru-RU": "Внутренняя ошибка."
  },
  "title_reset_password" : {
	 "en-US": "Reset Password",
	 "zh-CN": "重置密码",
	 "ja-JP": "パスワードをリセットする",
	 "de-DE": "Passwort zurücksetzen",
	 "ru-RU": "Сбросить пароль"
  },
  "title_sign_up" : {
	 "en-US": "Sign Up",
	 "zh-CN": "注册",
	 "ja-JP": "登録",
	 "de-DE": "Registrieren",
	 "ru-RU": "Регистрация"
  },
  "title_add_user": {
         "en-US": "Add User",
         "zh-CN": "新增用户",
	 "ja-JP": "ユーザを追加",
	 "de-DE": "Benutzer hinzufügen",
	 "ru-RU": "Добавить пользователя"
  },
  "registered_successfully": {
	 "en-US": "Signed up successfully.",
	 "zh-CN": "注册成功。",
	 "ja-JP": "登録しました。",
	 "de-DE": "Erfolgreich registriert.",
	 "ru-RU": "Регистрация прошла успешно."
  },
  "registered_failed" : {
	 "en-US": "Failed to sign up.",
	 "zh-CN": "注册失败。",
	 "ja-JP": "登録でませんでした。",
	 "de-DE": "Registrierung fehlgeschlagen.",
	 "ru-RU": "Ошибка регистрации."
  },
  "added_user_successfully": {
         "en-US": "Added user successfully.",
         "zh-CN": "新增用户成功。",
	 "ja-JP": "ユーザを追加しました。",
	 "de-DE": "Benutzer erfolgreich erstellt.",
	 "ru-RU": "Пользователь успешно добавлен."
  },
  "added_user_failed": {
         "en-US": "Adding user failed.",
         "zh-CN": "新增用户失败。",
	 "ja-JP": "ユーザを追加できませんでした。",
	 "de-DE": "Benutzer erstellen fehlgeschlagen.",
	 "ru-RU": "Ошибка добавления пользователя."
  },
  "projects":  {
	 "en-US": "Projects",
	 "zh-CN": "项目",
	 "ja-JP": "プロジェクト",
	 "de-DE": "Projekte",
	 "ru-RU": "Проекты"
  },
  "repositories" : {
	 "en-US": "Repositories",
	 "zh-CN": "镜像仓库",
	 "ja-JP": "リポジトリ",
	 "de-DE": "Repositories",
	 "ru-RU": "Репозитории"
  },
  "no_repo_exists"  : {
         "en-US": "No repositories found, please use 'docker push' to upload images.",
	 "zh-CN": "未发现镜像，请用‘docker push’命令上传镜像。",
	 "ja-JP": "イメージが見つかりませんでした。’docker push’を利用しイメージをアップロードしてください。",
	 "de-DE": "Keine Repositories gefunden, bitte benutzen Sie 'docker push' um ein Image hochzuladen.",
	 "ru-RU": "Репозитории не найдены, используйте команду 'docker push' для добавления образов."
  },
  "tag" : {
         "en-US": "Tag",
	 "zh-CN": "标签", 
	 "ja-JP": "タグ",
	 "de-DE": "Tag",
	 "ru-RU": "Метка"
  },
  "pull_command": {
	 "en-US": "Pull Command",
	 "zh-CN": "Pull 命令",
	 "ja-JP": "Pull コマンド",
	 "de-DE": "Pull Befehl",
	 "ru-RU": "Команда для скачивания образа"
  },
  "image_details" : {
	 "en-US": "Image Details",
	 "zh-CN": "镜像详细信息",
	 "ja-JP": "イメージ詳細",
	 "de-DE": "Image Details",
	 "ru-RU": "Информация об образе"
  },
  "add_members" : {
	 "en-US": "Add Member",
	 "zh-CN": "添加成员",
	 "ja-JP": "メンバーを追加する",
	 "de-DE": "Mitglied hinzufügen",
	 "ru-RU": "Добавить Участника"
  },
  "edit_members" : {
	 "en-US": "Edit Members",
	 "zh-CN": "编辑成员",
	 "ja-JP": "メンバーを編集する",
	 "de-DE": "Mitglieder bearbeiten",
	 "ru-RU": "Редактировать Участников"
  },
  "add_member_failed" : {
	 "en-US": "Adding Member Failed",
	 "zh-CN": "添加成员失败",
	 "ja-JP": "メンバーを追加できません出した",
	 "de-DE": "Mitglied hinzufügen fehlgeschlagen",
	 "ru-RU": "Ошибка при добавлении нового участника"
  },
  "please_input_username" : {
	 "en-US": "Please input a username.",
	 "zh-CN": "请输入用户名。",
	 "ja-JP": "ユーザ名を入力してください。",
	 "de-DE": "Bitte geben Sie einen Benutzernamen ein.",
	 "ru-RU": "Пожалуйста, введите имя пользователя."
  },
  "please_assign_a_role_to_user" : {
	 "en-US": "Please assign a role to the user.",
	 "zh-CN": "请为用户分配角色。",
	 "ja-JP": "ユーザーに役割を割り当てるしてください。",
	 "de-DE": "Bitte weisen Sie dem Benutzer eine Rolle zu.",
	 "ru-RU": "Пожалуйста, назначьте роль пользователю."
  },
  "user_id_exists" : {
	 "en-US": "User is already a member.",
	 "zh-CN": "用户已经是成员。",
	 "ja-JP": "すでにメンバーに登録しました。",
	 "de-DE": "Benutzer ist bereits Mitglied.",
	 "ru-RU": "Пользователь уже является участником."
  },
  "user_id_does_not_exist" : {
	 "en-US": "User does not exist.",
	 "zh-CN": "不存在此用户。",
	 "ja-JP": "ユーザが見つかりませんでした。",
	 "de-DE": "Benutzer existiert nicht.",
	 "ru-RU": "Пользователя с таким именем не существует."
  },
  "insufficient_privileges" : {
	 "en-US": "Insufficient privileges.",
	 "zh-CN": "权限不足。",
	 "ja-JP": "権限エラー。",
	 "de-DE": "Unzureichende Berechtigungen.",
	 "ru-RU": "Недостаточно прав."
  },
  "operation_failed" : {
	 "en-US": "Operation Failed",
	 "zh-CN": "操作失败",
	 "ja-JP": "操作に失敗しました。",
	 "de-DE": "Befehl fehlgeschlagen",
	 "ru-RU": "Ошибка при выполнении данной операции"
  },
  "button_on" : {
         "en-US": "On",
	 "zh-CN": "打开",
	 "ja-JP": "オン",
	 "de-DE": "An",
	 "ru-RU": "Вкл."
  },
  "button_off" : {
         "en-US": "Off",
	 "zh-CN": "关闭",
	 "ja-JP": "オフ",
	 "de-DE": "Aus",
	 "ru-RU": "Откл."
  }
};
