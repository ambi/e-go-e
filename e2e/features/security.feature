# language: en
@security
Feature: ログインフォームの CSRF 対策
  正しい認証情報でも、有効な CSRF トークンがなければログインできない。

  Scenario Outline: <条件>ではログインできない
    Given ログイン画面を開いている
    When <条件>で正しい認証情報を送信する
    Then フォームの送信が拒否される
    And マイページを開くとログイン画面へ戻される

    Examples:
      | 条件                 |
      | CSRFトークンなし     |
      | 不正なCSRFトークン   |
