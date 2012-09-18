require 'sinatra'

post '/' do
  p request
end

Sinatra::Application.run!
