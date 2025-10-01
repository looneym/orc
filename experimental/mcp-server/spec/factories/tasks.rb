FactoryBot.define do
  factory :task do
    title { "MyString" }
    description { "MyText" }
    status { "MyString" }
    priority { "MyString" }
    worktree { nil }
    created_by { "MyString" }
    assigned_agent { "MyString" }
  end
end
