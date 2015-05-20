require 'haml'

def partial(name, locals = {})
  source = File.read(File.join('frontend', 'src', 'partials', name.to_s + '.haml'))
  engine = Haml::Engine.new(source)
  engine.render(binding, locals: locals)
end
