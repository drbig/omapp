require 'haml'

def partial(name, locals = {})
  source = File.read(File.join('frontend', 'src', 'partials', name.to_s + '.haml'))
  engine = Haml::Engine.new(source)
  engine.render(binding, locals: locals)
end

def js(name)
  File.read(File.join('frontend', 'build', name.to_s + '.min.js'))
end

def css(name)
  File.read(File.join('frontend', 'build', name.to_s + '.css'))
end
