FactoryBot.define do
  factory :task_history do
    task { nil }
    action { "MyString" }
    old_value { "MyString" }
    new_value { "MyString" }
    notes { "MyText" }
    agent_id { "MyString" }
  end
end
