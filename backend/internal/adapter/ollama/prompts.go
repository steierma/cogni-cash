package llm

const defaultSinglePromptTemplate = `Categorize the following invoice text. 
Use EXACTLY ONE category from: [{{CATEGORIES}}].
Return ONLY a valid JSON object. Do not include explanations.

JSON Schema:
{"category_name": "string", "vendor_name": "string", "amount": 12.34, "currency": "string (ISO 4217, e.g. {{DEFAULT_CURRENCY}})", "invoice_date": "YYYY-MM-DD", "description": "string"}

INSTRUCTION:
Extract the actual currency from the document. If no currency is found, use {{DEFAULT_CURRENCY}}.

TEXT:
{{TEXT}}`

const defaultBatchPromptTemplate = `Categorize these transactions using ONLY: [{{CATEGORIES}}].
Return ONLY a valid JSON array of objects.
Each object MUST have "hash" and "category".

Here are some examples of past categorizations for reference:
{{EXAMPLES}}

DATA TO CATEGORIZE:
{{DATA}}`

const defaultStatementPromptTemplate = `Extract bank statement details from the following text.
Return ONLY a valid JSON object. Do not include explanations or markdown formatting outside of the JSON block.

JSON Schema:
{
  "account_holder": "string",
  "iban": "string",
  "currency": "string (ISO 4217, e.g. {{DEFAULT_CURRENCY}})",
  "statement_date": "YYYY-MM-DD",
  "statement_no": 123,
  "new_balance": 1234.56,
  "transactions": [
    {
      "booking_date": "YYYY-MM-DD",
      "amount": -12.34,
      "description": "string",
      "reference": "string",
      "counterparty_name": "string",
      "counterparty_iban": "string",
      "bank_transaction_code": "string",
      "mandate_reference": "string"
    }
  ]
}

INSTRUCTION:
Extract the actual currency from the bank statement. If no currency is found, use {{DEFAULT_CURRENCY}}.

TEXT:
{{TEXT}}`

const defaultPayslipPromptTemplate = `Role: You are a precise financial data extraction system.
Task: Extract payroll information from the provided payslip and map it strictly to the JSON schema below.

Strict Extraction Rules:

No Hallucinations: Extract values exactly as they are represented. Do not calculate, guess, or infer numbers.

Missing Data: If a value is not explicitly found in the text, you must return null for that field.

Number Formatting: Convert localized number formats (e.g., 1.234,56 or 1,234.56) into standard float values (e.g., 1234.56) without thousands separators.

Date Mapping: Convert month names found in the text into their corresponding integer (e.g., "January" / "Januar" = 1, "May" / "Mai" = 5).

Output Constraint: Return ONLY raw, valid JSON. Do not wrap the output in markdown code blocks (do not use ` + "```" + `json). Do not include any conversational text, explanations, or formatting outside the JSON object.

JSON Schema Definition:
{
  "period_month_num": "integer (1-12)",
  "period_year": "integer (YYYY)",
  "employer_name": "string",
  "tax_class": "string",
  "tax_id": "string",
  "gross_pay": "float",
  "net_pay": "float",
  "payout_amount": "float",
  "custom_deductions": "float or null",
  "bonuses": [{"description": "string", "amount": "float"}]
}

Source Text:
{{TEXT}}`

const defaultDocumentPromptTemplate = `Analyze the provided document and perform classification, metadata extraction, and a short summary.
Return ONLY a valid JSON object. Do not include explanations or markdown formatting outside of the JSON block.

Classification: Choose the most appropriate document type from: [tax_certificate, receipt, contract, other].

Metadata: Extract key information (e.g., date, amount, vendor, employer, reference numbers, etc.) as a JSON object.

Summary: Provide a one-sentence summary of the document.

JSON Schema:
{
  "document_type": "string",
  "metadata": {
    "key": "value"
  },
  "summary": "string"
}

DOCUMENT TEXT:
{{TEXT}}`

const defaultSubscriptionPromptTemplate = `Analyze the merchant and transaction descriptions to enrich subscription details.
Return ONLY a valid JSON object. Do not include explanations or markdown formatting outside of the JSON block.

Merchant: {{MERCHANT}}
Language: {{LANGUAGE}}
Transactions:
{{TRANSACTIONS}}

JSON Schema:
{
  "merchant_name": "string (cleaned/formal name)",
  "customer_number": "string (if found in transactions)",
  "contact_email": "string (support or billing email if known or found)",
  "contact_phone": "string (support phone if known or found)",
  "contact_website": "string (official website URL)",
  "support_url": "string (direct link to help/support page)",
  "cancellation_url": "string (direct link to cancellation or account management page)",
  "is_trial": "boolean (true if these transactions look like a free trial phase)",
  "notes": "string (short summary of the service in the requested language)"
}

If a field is unknown, return an empty string or null for boolean.`

const defaultVerificationPromptTemplate = `Analyze the merchant and transaction details to determine if it is a recurring subscription service.
A subscription service is something like Netflix, Rent, Insurance, Gym, Software (SaaS), etc.
One-off grocery purchases, random ATM withdrawals, or peer-to-peer transfers are NOT subscriptions.

Merchant: {{MERCHANT}}
Amount: {{AMOUNT}} {{CURRENCY}}
Billing Cycle: {{CYCLE}}

Return ONLY a valid JSON object. Do not include explanations.
JSON Schema:
{"is_subscription": boolean, "reason": "short explanation"}
`

const defaultCancellationLetterPromptTemplate = `Draft a formal cancellation letter for the following subscription.
The letter should be professional and include all necessary details for the merchant to identify the contract.
Return ONLY a valid JSON object. Do not include explanations or markdown formatting outside of the JSON block.

User: {{USER_NAME}} <{{USER_EMAIL}}>
Merchant: {{MERCHANT}}
Customer Number: {{CUSTOMER_NUMBER}}
End Date: {{END_DATE}}
Notice Period: {{NOTICE_PERIOD}} days
Language: {{LANGUAGE}}

JSON Schema:
{
  "subject": "string",
  "body": "string"
}

Draft the letter in the requested language ({{LANGUAGE}}).`
