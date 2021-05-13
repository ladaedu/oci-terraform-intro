output "email_suggestion" {
    value = <<EOF
Dobrý den,
bylo Vám vytvořeno konto v Oracle Cloud Infrastructure (OCI), s dočasným heslem, které si musíte při prvním přihlášení změnit.

Údaje pro přihlášení:
Login URL:  https://console.eu-frankfurt-1.oraclecloud.com/?tenant=${var.tenancy_name}
User Name:  `${var.email}`
Password:   `${oci_identity_ui_password.reset_password.password}`

Pro přihlášení jděte na Login URL a tam použijte sekci "Oracle Cloud Infrastructure Direct Sign-In".

Byl jste přidán do skupiny "${var.group}".
OCI vám poslal email na potvrzení Vaší emailové adresy - abyste si případně mohl změnit heslo, když ho zapomenete. Na odkaz v tom emailu doporučuji kliknout, až poté, co provedete první přihlášení.

Pokud chcete být připraven na terraform, doporučuji si vygenerovat API Key ve Vašem profilu.

Kdyby něco, napište ;-)

EOF
    description = "Email"
}

